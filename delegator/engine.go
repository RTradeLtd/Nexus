package delegator

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/RTradeLtd/Nexus/config"
	"github.com/RTradeLtd/Nexus/ipfs"
	"github.com/RTradeLtd/Nexus/log"
	"github.com/RTradeLtd/Nexus/network"
	"github.com/RTradeLtd/Nexus/registry"
	"github.com/RTradeLtd/Nexus/temporal"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// Engine manages request delegation
type Engine struct {
	l     *zap.SugaredLogger
	reg   *registry.NodeRegistry
	cache *cache

	access   temporal.AccessChecker
	networks temporal.PrivateNetworks

	timeout   time.Duration
	keyLookup jwt.Keyfunc
	version   string
}

// New instantiates a new delegator engine
func New(l *zap.SugaredLogger, version string, timeout time.Duration, jwtKey []byte,
	reg *registry.NodeRegistry, access temporal.AccessChecker, networks temporal.PrivateNetworks) *Engine {

	return &Engine{
		l:     l.Named("delegator"),
		reg:   reg,
		cache: newCache(30*time.Minute, 30*time.Minute),

		access:   access,
		networks: networks,

		timeout: timeout,
		version: version,
		keyLookup: func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
	}
}

// Run spins up a server that listens for requests and proxies them appropriately
func (e *Engine) Run(ctx context.Context, opts config.Delegator) error {
	var r = chi.NewRouter()

	// mount middleware
	r.Use(
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler,
		middleware.RequestID,
		middleware.RealIP,
		log.NewMiddleware(e.l.Named("requests")),
		middleware.Recoverer,
	)

	// register endpoints
	r.HandleFunc("/status", e.Status)
	r.Route(fmt.Sprintf("/network/{%s}", keyNetwork), func(r chi.Router) {
		r.Use(e.NetworkContext)
		r.HandleFunc("/status", e.NetworkStatus)
		r.Route(fmt.Sprintf("/{%s}", keyFeature), func(r chi.Router) {
			r.HandleFunc("/*", e.Redirect)
		})
	})

	// set up server
	var srv = &http.Server{
		Handler: r,

		Addr:         opts.Host + ":" + opts.Port,
		WriteTimeout: e.timeout,
		ReadTimeout:  e.timeout,
	}

	// handle shutdown
	go func() {
		for {
			select {
			case <-ctx.Done():
				shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				if err := srv.Shutdown(shutdown); err != nil {
					e.l.Warnw("error encountered during shutdown", "error", err.Error())
				}
			}
		}
	}()

	// go!
	if opts.TLS.CertPath != "" {
		if err := srv.ListenAndServeTLS(opts.TLS.CertPath, opts.TLS.KeyPath); err != nil && err != http.ErrServerClosed {
			e.l.Errorw("error encountered - service stopped", "error", err)
			return err
		}
	} else {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.l.Errorw("error encountered - service stopped", "error", err)
			return err
		}
	}

	return nil
}

// NetworkContext creates a handler that injects relevant network context into
// all incoming requests through URL parameters
func (e *Engine) NetworkContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id = chi.URLParam(r, string(keyNetwork))
		n, err := e.reg.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), keyNetwork, &n),
		))
	})
}

// Redirect manages request redirects
func (e *Engine) Redirect(w http.ResponseWriter, r *http.Request) {
	// retrieve network
	n, ok := r.Context().Value(keyNetwork).(*ipfs.NodeInfo)
	if !ok || n == nil {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	// retrieve requested feature
	var feature string
	if feature = chi.URLParam(r, string(keyFeature)); feature == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// set target port based on feature
	var port string
	switch feature {
	case "swarm":
		port = n.Ports.Swarm
	case "api":
		user, err := getUserFromJWT(r, e.keyLookup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if ok, err := e.access.CheckIfUserHasAccessToNetwork(user, n.NetworkID); err != nil {
			http.Error(w, "failed to find user", http.StatusNotFound)
			return
		} else if !ok {
			http.Error(w, "user not authorized", http.StatusForbidden)
			return
		}
		port = n.Ports.API
	case "gateway":
		if entry, err := e.networks.GetNetworkByName(n.NetworkID); err != nil {
			http.Error(w, "failed to find network", http.StatusNotFound)
			return
		} else if !entry.GatewayPublic {
			http.Error(w, "failed to find network gateway", http.StatusNotFound)
			return
		}
		port = n.Ports.Gateway
	default:
		http.Error(w, fmt.Sprintf("invalid feature '%s'", feature), http.StatusBadRequest)
		return
	}

	// set up target
	var protocol string
	if r.URL.Scheme != "" {
		protocol = r.URL.Scheme + "://"
	} else {
		protocol = "http://"
	}

	var (
		address = fmt.Sprintf("%s:%s", network.Private, port)
		target  = fmt.Sprintf("%s%s%s", protocol, address, r.RequestURI)
	)

	url, err := url.Parse(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set up forwarder, retrieving from cache if available, otherwise set up new
	var proxy *httputil.ReverseProxy
	if proxy = e.cache.Get(fmt.Sprintf("%s-%s", n.NetworkID, feature)); proxy == nil {
		proxy = newProxy(feature, url, e.l)
		e.cache.Cache(fmt.Sprintf("%s-%s", n.NetworkID, feature), proxy)
	}

	// serve proxy request
	proxy.ServeHTTP(w, r)
}

// Status reports on proxy status
func (e *Engine) Status(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]string{
		"status":  "online",
		"version": e.version,
	})
}

// NetworkStatus reports on the status of a network
func (e *Engine) NetworkStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Context().Value(keyNetwork).(*ipfs.NodeInfo); !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]string{
		"status": "registered",
	})
}

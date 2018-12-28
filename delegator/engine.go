package delegator

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// Engine manages request delegation
type Engine struct {
	l   *zap.SugaredLogger
	reg *registry.NodeRegistry
	net *http.Client

	timeout time.Duration
	version string
}

// New instantiates a new delegator engine
func New(l *zap.SugaredLogger, version string, timeout time.Duration, reg *registry.NodeRegistry) *Engine {
	return &Engine{
		l:   l.Named("delegator"),
		reg: reg,
		net: http.DefaultClient,

		timeout: timeout,
		version: version,
	}
}

// Run spins up a server that listens for requests and proxies them appropriately
func (e *Engine) Run(ctx context.Context, opts config.Proxy) error {
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
		newLoggerMiddleware(e.l.Named("requests")),
		middleware.Recoverer,
	)

	// register endpoints
	r.HandleFunc("/status", e.Status)
	r.Route(fmt.Sprintf("/network/{%s}", keyNetwork), func(r chi.Router) {
		r.Use(e.Context)
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

// Context injects relevant context into all requests
func (e *Engine) Context(next http.Handler) http.Handler {
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
	var f = chi.URLParam(r, string(keyFeature))
	if f == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// set target port based on feature
	var port string
	switch f {
	case "api":
		port = n.Ports.API
	case "swarm":
		port = n.Ports.Swarm
	default:
		http.Error(w, fmt.Sprintf("invalid feature '%s'", f), http.StatusBadRequest)
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
		address = fmt.Sprintf("%s:%s", "0.0.0.0", port)
		target  = fmt.Sprintf("%s%s%s", protocol, address, r.RequestURI)
	)

	// set up forwarder - TODO: cache
	url, err := url.Parse(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxy := newProxy(f, url, e.l)

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

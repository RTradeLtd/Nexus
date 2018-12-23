package delegator

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// Engine manages request delegation
type Engine struct {
	l   *zap.SugaredLogger
	reg *registry.NodeRegistry
	net *http.Client

	timeout time.Duration
}

// New instantiates a new delegator engine
func New(l *zap.SugaredLogger, timeout time.Duration, reg *registry.NodeRegistry) *Engine {
	return &Engine{
		l:   l.Named("delegator"),
		reg: reg,
		net: http.DefaultClient,

		timeout: timeout,
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
		newLoggerMiddleware(e.l),
		middleware.Recoverer,
		middleware.Timeout(e.timeout),
	)

	// register endpoints
	r.Route(fmt.Sprintf("/networks/{%s}", keyNetwork), func(r chi.Router) {
		r.Use(e.Context)
		r.HandleFunc(fmt.Sprintf("/{%s}", keyFeature), e.Redirect)
	})

	// set up server
	var srv = &http.Server{Addr: opts.Host + ":" + opts.Port, Handler: r}

	// handle shutdown
	go func() error {
		for {
			select {
			case <-ctx.Done():
				shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				if err := srv.Shutdown(shutdown); err != nil {
					e.l.Warnw("error encountered during shutdown", "error", err.Error())
					return err
				}
				return nil
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
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), keyNetwork, n),
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
	f, ok := r.Context().Value(keyFeature).(string)
	if !ok || f == "" {
		http.Error(w, http.StatusText(422), 422)
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
	}

	// set target
	var (
		protocol = r.URL.Scheme
		address  = fmt.Sprintf("%s:%s", "0.0.0.0", port)
		target   = fmt.Sprintf("%s://%s%s", protocol, address, r.RequestURI)
	)

	// read request for forwarding
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// set up proxy
	proxy, err := http.NewRequest(r.Method, target, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxy.Header = r.Header

	// execute forward
	resp, err := e.net.Do(proxy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
}

package delegator

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Engine manages request delegation
type Engine struct {
	l   *zap.SugaredLogger
	reg *registry.NodeRegistry
	net *http.Client

	timeout time.Duration
}

// RunOpts declares options for running the delegator engine
type RunOpts struct {
	certpath string
	keypath  string
}

// New instantiates a new delegator engine
func New(l *zap.SugaredLogger, timeout time.Duration, reg *registry.NodeRegistry) (*Engine, error) {
	l = l.Named("delegator")
	return &Engine{
		l:   l,
		reg: reg,
		net: http.DefaultClient,

		timeout: timeout,
	}, nil
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

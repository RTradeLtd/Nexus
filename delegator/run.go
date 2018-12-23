package delegator

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

// Run spins up a server that listens for requests and proxies them appropriately
func Run(e *Engine, addr string, tls *RunOpts) error {
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

	// spin up server
	if tls != nil {
		return http.ListenAndServeTLS(addr, tls.certpath, tls.keypath, r)
	}
	return http.ListenAndServe(addr, r)
}

package delegator

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

func newProxy(feature string, target *url.URL, l *zap.SugaredLogger) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// remove delgator-specific leading elements, e.g. /networks/test_network/api,
			// and accomodate for specific cases
			switch feature {
			case "api":
				req.URL.Path = "/api" + stripLeadingSegments(req.URL.Path)
			default:
				req.URL.Path = stripLeadingSegments(req.URL.Path)
			}

			// set other URL properties
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host

			l.Debugw("forwarded request",
				"path", req.URL.Path,
				"url", req.URL)
		},
	}
}

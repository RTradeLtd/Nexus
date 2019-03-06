package delegator

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

func newProxy(feature string, target *url.URL, l *zap.SugaredLogger, direct bool) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// if set up as an indirect proxy, we need to remove delgator-specific
			// leading elements, e.g. /networks/test_network/api, from the path and
			// accomodate for specific cases
			if !direct {
				switch feature {
				case "api":
					req.URL.Path = "/api" + stripLeadingSegments(req.URL.Path)
				default:
					req.URL.Path = stripLeadingSegments(req.URL.Path)
				}
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

func validateFeature(feature string) bool {
	switch feature {
	case "api":
		fallthrough
	case "swarm":
		fallthrough
	case "gateway":
		return true
	default:
		return false
	}
}

func stripLeadingSegments(path string) string {
	var expected = 5
	var parts = strings.SplitN(path, "/", expected)
	if len(parts) == expected {
		return "/" + parts[expected-1]
	}
	return path
}

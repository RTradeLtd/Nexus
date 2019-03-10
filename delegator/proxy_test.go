package delegator

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"go.uber.org/zap/zaptest"
)

func Test_newProxy(t *testing.T) {
	var testBytes = []byte("{\"hello\":\"world\"}")
	type args struct {
		feature string
		target  *url.URL
		direct  bool
	}
	type request struct {
		method  string
		address string
		body    io.Reader
	}
	tests := []struct {
		name string
		args args
		req  request
		want *url.URL
	}{
		{"api redirect with GET",
			args{"api", &url.URL{Scheme: "http", Host: "127.0.0.1"}, false},
			request{"GET", "http://127.0.0.1/networks/blah/api/v0/blah", nil},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/api/v0/blah"}},
		{"api redirect with POST",
			args{"api", &url.URL{Scheme: "http", Host: "127.0.0.1"}, false},
			request{"POST", "http://127.0.0.1/networks/blah/api/v0/blah", bytes.NewReader(testBytes)},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/api/v0/blah"}},
		{"indirect redirect",
			args{"swarm", &url.URL{Scheme: "http", Host: "127.0.0.1"}, false},
			request{"GET", "http://127.0.0.1/networks/blah/swarm/blah", nil},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/blah"}},
		{"direct redirect",
			args{"swarm", &url.URL{Scheme: "http", Host: "127.0.0.1"}, true},
			request{"GET", "https://node.gateway.temporal.cloud/blah", nil},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/blah"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l = zaptest.NewLogger(t).Sugar()
			var got = newProxy(tt.args.feature, tt.args.target, l, tt.args.direct)

			// test proxy Director
			var r = httptest.NewRequest(tt.req.method, tt.req.address, tt.req.body)
			got.Director(r)

			// check target
			if !reflect.DeepEqual(r.URL, tt.want) {
				t.Errorf("expected URL '%v', got '%v'", tt.want, r.URL)
			}

			// check body
			if tt.req.body != nil {
				found, _ := ioutil.ReadAll(r.Body)
				if !reflect.DeepEqual(found, testBytes) {
					t.Errorf("expected body '%v', got '%v'", testBytes, found)
				}
			}
		})
	}
}

func Test_stripLeadingPaths(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"remove prefix", args{"/networks/test_network/api/version"}, "/version"},
		{"no prefix removed", args{"/api/version"}, "/api/version"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripLeadingSegments(tt.args.path); got != tt.want {
				t.Errorf("stripLeadingPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

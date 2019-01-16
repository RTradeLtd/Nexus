package delegator

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/RTradeLtd/Nexus/log"
)

func Test_newProxy(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	var testBytes = []byte("{\"hello\":\"world\"}")
	type args struct {
		feature string
		target  *url.URL
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
			args{"api", &url.URL{Scheme: "http", Host: "127.0.0.1"}},
			request{"GET", "http://127.0.0.1/networks/blah/api/v0/blah", nil},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/api/v0/blah"}},
		{"api redirect with POST",
			args{"api", &url.URL{Scheme: "http", Host: "127.0.0.1"}},
			request{"POST", "http://127.0.0.1/networks/blah/api/v0/blah", bytes.NewReader(testBytes)},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/api/v0/blah"}},
		{"default redirect",
			args{"swarm", &url.URL{Scheme: "http", Host: "127.0.0.1"}},
			request{"GET", "http://127.0.0.1/networks/blah/swarm/blah", nil},
			&url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/blah"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got = newProxy(tt.args.feature, tt.args.target, l)

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

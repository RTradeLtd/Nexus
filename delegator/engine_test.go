package delegator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RTradeLtd/database/models"

	"github.com/RTradeLtd/Nexus/config"
	"github.com/RTradeLtd/Nexus/ipfs"
	"github.com/RTradeLtd/Nexus/log"
	"github.com/RTradeLtd/Nexus/registry"
	"github.com/RTradeLtd/Nexus/temporal/mock"
	"github.com/go-chi/chi"
)

func TestEngine_Run(t *testing.T) {
	// claim port for testing unavailable port
	if port, err := net.Listen("tcp", "127.0.0.1:69"); err != nil && port != nil {
		defer port.Close()
	}

	var l, _ = log.NewLogger("", true)
	type args struct {
		opts config.Delegator
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no cert",
			args{config.Delegator{
				Host: "127.0.0.1",
				Port: "",
			}},
			false},
		{"invalid port",
			args{config.Delegator{
				Host: "127.0.0.1",
				Port: "69",
			}},
			true},
		{"invalid cert",
			args{config.Delegator{
				Host: "127.0.0.1",
				Port: "",
				TLS:  config.TLS{CertPath: "../README.md"},
			}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var networks = &mock.FakePrivateNetworks{}
			var e = New(l, EngineOpts{"test", true, "domain.com", time.Minute, []byte("hello")}, nil, networks)
			var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := e.Run(ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Engine.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_NetworkPathContext(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	type args struct {
		nodeName string
		key      contextKey
		target   string
	}
	tests := []struct {
		name     string
		args     args
		wantNode bool
		wantCode int
	}{
		{"non existent node", args{"hello", keyNetwork, "bye"}, false, http.StatusNotFound},
		{"invalid key", args{"hello", keyFeature, "hello"}, false, http.StatusNotFound},
		{"find node", args{"hello", keyNetwork, "hello"}, true, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var networks = &mock.FakePrivateNetworks{}
			var e = New(l, EngineOpts{"test", true, "", time.Second, []byte("hello")},
				registry.New(l, config.New().Ports, &ipfs.NodeInfo{
					NetworkID: tt.args.nodeName,
				}), networks)

			// set up route context and request
			var route = chi.NewRouteContext()
			route.URLParams.Add(string(tt.args.key), tt.args.target)
			var req = httptest.NewRequest("GET", "/", nil).
				WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, route))

			// test handler
			var rec = httptest.NewRecorder()
			var ok bool
			e.NetworkPathContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				// check for node in context
				var n *ipfs.NodeInfo
				n, ok = r.Context().Value(keyNetwork).(*ipfs.NodeInfo)
				if (!ok || n == nil) && tt.wantNode {
					t.Errorf("expected ipfs node, found '%v'", r.Context().Value(keyNetwork))
				}
				if tt.wantNode && n.NetworkID != tt.args.nodeName {
					t.Errorf("expected node named '%s', found '%s'", tt.args.nodeName, n.NetworkID)
				}
				return
			})).ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("expected status '%d', found '%d'", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestEngine_FeaturePathContext(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	type args struct {
		key    contextKey
		target string
	}
	tests := []struct {
		name        string
		args        args
		wantFeature string
		wantCode    int
	}{
		{"invalid feature", args{keyFeature, "bye"}, "", http.StatusBadRequest},
		{"invalid key", args{keyNetwork, "hello"}, "", http.StatusBadRequest},
		{"ok feature", args{keyFeature, "swarm"}, "swarm", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var networks = &mock.FakePrivateNetworks{}
			var e = New(l, EngineOpts{"test", true, "", time.Second, []byte("hello")},
				registry.New(l, config.New().Ports), networks)

			// set up route context and request
			var route = chi.NewRouteContext()
			route.URLParams.Add(string(tt.args.key), tt.args.target)
			var req = httptest.NewRequest("GET", "/", nil).
				WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, route))

			// test handler
			var rec = httptest.NewRecorder()
			var ok bool
			e.FeaturePathContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				// check for feature in context
				var feature string
				feature, ok = r.Context().Value(keyFeature).(string)
				if (!ok) && tt.wantFeature != "" {
					t.Errorf("expected feature, found '%v'", r.Context().Value(keyFeature))
				}
				if tt.wantFeature != "" && tt.wantFeature != feature {
					t.Errorf("expected feature '%s', found '%s'", tt.wantFeature, feature)
				}
				return
			})).ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("expected status '%d', found '%d'", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestEngine_NetworkSubdomainContext(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	type args struct {
		nodeName string
		domain   string
	}
	tests := []struct {
		name        string
		args        args
		wantNode    bool
		wantFeature string
		wantCode    int
	}{
		{"non existent node", args{"hello", "http://blah.api.bye/"}, false, "", http.StatusNotFound},
		{"invalid feature", args{"hello", "http://hello.asdf.bye/"}, false, "", http.StatusBadRequest},
		{"find node", args{"hello", "http://hello.api.asdf/"}, true, "api", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var networks = &mock.FakePrivateNetworks{}
			var e = New(l, EngineOpts{"test", true, "", time.Second, []byte("hello")},
				registry.New(l, config.New().Ports, &ipfs.NodeInfo{
					NetworkID: tt.args.nodeName,
				}), networks)

			// construct request
			var req = httptest.NewRequest("GET", tt.args.domain, nil)

			// test handler
			var rec = httptest.NewRecorder()
			var ok bool
			e.NetworkAndFeatureSubdomainContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// check for node in context
				var n *ipfs.NodeInfo
				n, ok = r.Context().Value(keyNetwork).(*ipfs.NodeInfo)
				if (!ok || n == nil) && tt.wantNode {
					t.Errorf("expected ipfs node, found '%v'", r.Context().Value(keyNetwork))
				}
				if tt.wantNode && n.NetworkID != tt.args.nodeName {
					t.Errorf("expected node named '%s', found '%s'", tt.args.nodeName, n.NetworkID)
				}

				// check for feature in context
				var feature string
				feature, ok = r.Context().Value(keyFeature).(string)
				if (!ok || n == nil) && tt.wantFeature != "" {
					t.Errorf("expected feature, found '%v'", r.Context().Value(keyFeature))
				}
				if tt.wantFeature != "" && tt.wantFeature != feature {
					t.Errorf("expected feature '%s', found '%s'", tt.wantFeature, feature)
				}
				return
			})).ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("expected status '%d', found '%d'", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestEngine_Redirect(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	type fields struct {
		node       *ipfs.NodeInfo
		network    *models.HostedNetwork
		networkErr error
	}
	type args struct {
		token string
		route map[contextKey]string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int
	}{
		{"no network", fields{nil, nil, nil}, args{"", nil}, http.StatusUnprocessableEntity},
		{"no feature", fields{&ipfs.NodeInfo{}, nil, nil}, args{"", nil}, http.StatusBadRequest},
		{"bad feature",
			fields{&ipfs.NodeInfo{}, nil, nil},
			args{"", map[contextKey]string{keyFeature: "bobheadxi"}},
			http.StatusBadRequest},
		{"OK: swarm",
			fields{&ipfs.NodeInfo{Ports: ipfs.NodePorts{Swarm: "5000"}}, nil, nil},
			args{"", map[contextKey]string{keyFeature: "swarm"}},
			http.StatusBadGateway}, // badgateway because proxy points to nothing
		{"api + bad token",
			fields{&ipfs.NodeInfo{Ports: ipfs.NodePorts{API: "5000"}}, nil, nil},
			args{"asdf", map[contextKey]string{keyFeature: "api"}},
			http.StatusUnauthorized},
		{"api + good token + no database entry",
			fields{&ipfs.NodeInfo{Ports: ipfs.NodePorts{API: "5000"}}, nil, errors.New("oh")},
			args{validToken, map[contextKey]string{keyFeature: "api"}},
			http.StatusNotFound},
		{"api + good token + no authorization",
			fields{
				&ipfs.NodeInfo{Ports: ipfs.NodePorts{API: "5000"}},
				&models.HostedNetwork{},
				nil},
			args{validToken, map[contextKey]string{keyFeature: "api"}},
			http.StatusForbidden},
		{"api + good token + disallowed origin",
			fields{
				&ipfs.NodeInfo{Ports: ipfs.NodePorts{API: "5000"}},
				&models.HostedNetwork{
					Users:            []string{"testuser"},
					APIAllowedOrigin: "https://www.google.com",
				},
				nil},
			args{validToken, map[contextKey]string{keyFeature: "api"}},
			http.StatusBadGateway}, // badgateway because proxy points to nothing
		{"OK: api + good token + authorization",
			fields{
				&ipfs.NodeInfo{Ports: ipfs.NodePorts{API: "5000"}},
				&models.HostedNetwork{
					Users: []string{"testuser"},
				},
				nil},
			args{validToken, map[contextKey]string{keyFeature: "api"}},
			http.StatusBadGateway}, // badgateway because proxy points to nothing
		{"gateway + good token + no network",
			fields{
				&ipfs.NodeInfo{Ports: ipfs.NodePorts{Gateway: "5000"}},
				nil,
				errors.New("oh")},
			args{validToken, map[contextKey]string{keyFeature: "gateway"}},
			http.StatusNotFound},
		{"gateway + good token + not public",
			fields{
				&ipfs.NodeInfo{Ports: ipfs.NodePorts{Gateway: "5000"}},
				&models.HostedNetwork{
					GatewayPublic: false,
				},
				nil},
			args{validToken, map[contextKey]string{keyFeature: "gateway"}},
			http.StatusNotFound},
		{"OK: gateway + good token + public",
			fields{
				&ipfs.NodeInfo{Ports: ipfs.NodePorts{Gateway: "5000"}},
				&models.HostedNetwork{
					GatewayPublic: true,
				},
				nil},
			args{validToken, map[contextKey]string{keyFeature: "gateway"}},
			http.StatusBadGateway}, // badgateway because proxy points to nothing
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var networks = &mock.FakePrivateNetworks{}
			var e = New(l, EngineOpts{"test", true, "", time.Second, defaultTestKey},
				registry.New(l, config.New().Ports), networks)

			networks.GetNetworkByNameReturns(tt.fields.network, tt.fields.networkErr)

			var ctx = context.WithValue(
				context.Background(),
				keyNetwork, tt.fields.node)
			if tt.args.route != nil {
				for key, val := range tt.args.route {
					ctx = context.WithValue(ctx, key, val)
				}
			}

			var (
				req = httptest.NewRequest("GET", "/", nil).WithContext(ctx)
				rec = httptest.NewRecorder()
			)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.args.token))
			e.Redirect(rec, req)
			if rec.Code != tt.wantCode {
				t.Logf("received '%v'", rec.Result().Status)
				t.Errorf("expected status '%d', found '%d'", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestEngine_Status(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	var networks = &mock.FakePrivateNetworks{}
	var e = New(l,
		EngineOpts{"test", true, "", time.Second, []byte("hello")},
		registry.New(l, config.New().Ports),
		networks)
	var req = httptest.NewRequest("GET", "/", nil)
	var rec = httptest.NewRecorder()
	e.Status(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status '%d', found '%d'", http.StatusOK, rec.Code)
	}
}

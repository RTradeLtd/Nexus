package delegator

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"github.com/RTradeLtd/ipfs-orchestrator/temporal/mock"
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
			var checker = &mock.FakeAccessChecker{}
			var e = New(l, "test", time.Minute, []byte("hello"), nil, checker)
			var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := e.Run(ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Engine.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_NetworkContext(t *testing.T) {
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
			var checker = &mock.FakeAccessChecker{}
			var e = New(l, "test", time.Second, []byte("hello"),
				registry.New(l, config.New().Ports, &ipfs.NodeInfo{
					NetworkID: tt.args.nodeName,
				}), checker)
			// set up route context and request
			var route = chi.NewRouteContext()
			route.URLParams.Add(string(tt.args.key), tt.args.target)
			var req = httptest.NewRequest("GET", "/", nil).
				WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, route))

			// test handler
			var rec = httptest.NewRecorder()
			var n *ipfs.NodeInfo
			e.NetworkContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				var ok bool
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

func TestEngine_Status(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	var checker = &mock.FakeAccessChecker{}
	var e = New(l, "test", time.Second, []byte("hello"), registry.New(l, config.New().Ports), checker)
	var req = httptest.NewRequest("GET", "/", nil)
	var rec = httptest.NewRecorder()
	e.Status(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status '%d', found '%d'", http.StatusOK, rec.Code)
	}
}

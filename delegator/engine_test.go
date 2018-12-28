package delegator

import (
	"context"
	"testing"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
)

func TestEngine_Run(t *testing.T) {
	var l, _ = log.NewLogger("", true)
	type args struct {
		opts config.Proxy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no cert",
			args{config.Proxy{
				Host: "127.0.0.1",
				Port: "",
			}},
			false},
		{"invalid port",
			args{config.Proxy{
				Host: "127.0.0.1",
				Port: "7",
			}},
			true},
		{"invalid cert",
			args{config.Proxy{
				Host: "127.0.0.1",
				Port: "",
				TLS:  config.TLS{CertPath: "../README.md"},
			}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e = New(l, "test", time.Minute, nil)
			var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := e.Run(ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Engine.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

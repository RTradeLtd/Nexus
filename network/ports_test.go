package network

import (
	"net"
	"testing"
	"time"

	"github.com/RTradeLtd/Nexus/log"
)

func TestNewRegistry(t *testing.T) {
	l, _ := log.NewTestLogger()
	NewRegistry(l, "127.0.0.1", []string{"1234"})
	NewRegistry(l, "127.0.0.1", nil)
}

func TestRegistry_AssignPort(t *testing.T) {
	// lock a port for testing
	p1, _ := net.Listen("tcp", "127.0.0.1:9999")
	defer p1.Close()

	type fields struct {
		ports []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"nil ports", fields{nil}, "", true},
		{"no ports", fields{[]string{}}, "", true},
		{"no available port", fields{[]string{"9999"}}, "", true},
		{"available port", fields{[]string{"9998"}}, "9998", false},
		{"cache and try next port", fields{[]string{"9999", "9998"}}, "9998", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := log.NewTestLogger()
			reg := &Registry{l: l, host: "127.0.0.1", ports: tt.fields.ports,
				recent: newCache(5*time.Minute, 10*time.Minute)}
			defer reg.Close()
			got, err := reg.AssignPort()
			if (err != nil) != tt.wantErr {
				t.Errorf("Registry.AssignPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Registry.AssignPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

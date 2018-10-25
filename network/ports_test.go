package network

import (
	"net"
	"reflect"
	"testing"

	"github.com/RTradeLtd/ipfs-orchestrator/log"
)

func TestNewRegistry(t *testing.T) {
	l, _ := log.NewTestLogger()
	NewRegistry(l, "127.0.0.1", []string{"1234"})
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
		{"no ports", fields{[]string{}}, "", true},
		{"no available port", fields{[]string{"9999"}}, "", true},
		{"available port", fields{[]string{"9998"}}, "9998", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := log.NewTestLogger()
			reg := &Registry{l: l, host: "127.0.0.1", ports: tt.fields.ports}
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

func Test_parsePorts(t *testing.T) {
	type args struct {
		portRanges []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"bad input", args{nil}, []string{}},
		{"non-int port", args{[]string{"abcde"}}, []string{}},
		{"non-int port range", args{[]string{"abcde-fghijk"}}, []string{}},
		{"mixed non-int port range", args{[]string{"7000-abcde"}}, []string{}},
		{"bad range", args{[]string{"8000-7999"}}, []string{}},
		{"single port", args{[]string{"8000"}}, []string{"8000"}},
		{"multi port", args{[]string{"8000", "8001"}}, []string{"8000", "8001"}},
		{"port range", args{[]string{"8000-8003"}}, []string{"8000", "8001", "8002", "8003"}},
		{"multi port range", args{[]string{"8000-8001", "8002-8003"}}, []string{"8000", "8001", "8002", "8003"}},
		{"mix", args{[]string{"8000", "8002-8003"}}, []string{"8000", "8002", "8003"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePorts(tt.args.portRanges); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

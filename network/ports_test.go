package network

import (
	"net"
	"reflect"
	"testing"
)

func TestRegistry_lockPorts(t *testing.T) {
	type args struct {
		host       string
		portRanges []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"should lock single port", args{"127.0.0.1", []string{"8090"}}},
		{"should lock ports in range", args{"127.0.0.1", []string{"8090-8100"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry(tt.args.host, tt.args.portRanges)
			for p, lock := range reg.ports {
				if lock == nil {
					t.Logf("%s not locked", p)
					if _, err := net.Listen("tcp", tt.args.host+":"+p); err != nil {
						t.Errorf("%s should have been claimed", p)
					}
					continue
				}
				t.Logf("%s successfully locked at %s", p, lock.Addr().String())
				if _, err := net.Listen("tcp", tt.args.host+":"+p); err == nil {
					t.Errorf("%s was not successfully claimed", p)
				}
				lock.Close()
			}
		})
	}
}

func TestRegistry_AssignPort(t *testing.T) {
	// lock a port for testing
	p1, _ := net.Listen("tcp", "127.0.0.1:9999")
	defer p1.Close()

	type fields struct {
		ports map[string]net.Listener
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"no available port", fields{map[string]net.Listener{"9999": nil}}, "", true},
		{"available port", fields{map[string]net.Listener{"9999": p1}}, "9999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{ports: tt.fields.ports}
			got, err := reg.AssignPort()
			if (err != nil) != tt.wantErr {
				t.Errorf("Registry.AssignPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Registry.AssignPort() = %v, want %v", got, tt.want)
			}
			if !tt.wantErr && reg.ports[got] != nil {
				t.Errorf("Assigned port %s was not made unavailable", got)
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

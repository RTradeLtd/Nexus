package registry

import (
	"net"
	"reflect"
	"testing"
)

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

func Test_lockPorts(t *testing.T) {
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
		reg := make(map[string]net.Listener)
		t.Run(tt.name, func(t *testing.T) {
			lockPorts(tt.args.host, tt.args.portRanges, reg)
		})

		for p, lock := range reg {
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
	}
}

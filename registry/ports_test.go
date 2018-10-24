package registry

import (
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

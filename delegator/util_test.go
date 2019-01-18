package delegator

import "testing"

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

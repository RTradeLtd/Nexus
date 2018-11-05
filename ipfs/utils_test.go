package ipfs

import (
	"testing"
)

func Test_isNodeContainer(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"is node container", args{"ipfs-node1"}, true},
		{"is node container with default docker name", args{"/ipfs-node1"}, true},
		{"not node container", args{"abcde-node1"}, false},
		{"not node container", args{"abcdenode1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNodeContainer(tt.args.imageName); got != tt.want {
				t.Errorf("isNodeContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}

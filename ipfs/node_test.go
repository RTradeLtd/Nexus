package ipfs

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNode(t *testing.T) {
	// super simple tests to get coverage on node private data getters
	n := NodeInfo{}
	n.DockerID()
	n.ContainerName()
	n.DataDirectory()
	n.BootstrapPeers()
}

func Test_newNode(t *testing.T) {
	bl := []string{"1234"}
	b, _ := json.Marshal(&bl)
	type args struct {
		id         string
		name       string
		attributes map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    NodeInfo
		wantErr bool
	}{
		{"not an ipfs node",
			args{"1", "not-an-ipfs-node1", map[string]string{}},
			NodeInfo{dockerID: "1", containerName: "not-an-ipfs-node1"},
			true},
		{"default",
			args{"1", "ipfs-node1", map[string]string{}},
			NodeInfo{dockerID: "1", containerName: "ipfs-node1"},
			false},
		{"parse bootstrap",
			args{"1", "ipfs-node1", map[string]string{"bootstrap_peers": string(b)}},
			NodeInfo{dockerID: "1", containerName: "ipfs-node1", bootstrapPeers: bl},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newNode(tt.args.id, tt.args.name, tt.args.attributes)
			if (err != nil) != tt.wantErr {
				t.Errorf("newNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

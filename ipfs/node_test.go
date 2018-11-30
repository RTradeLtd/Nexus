package ipfs

import (
	"encoding/json"
	"reflect"
	"testing"
)

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
			NodeInfo{DockerID: "1", ContainerName: "not-an-ipfs-node1"},
			true},
		{"default",
			args{"1", "ipfs-node1", map[string]string{}},
			NodeInfo{DockerID: "1", ContainerName: "ipfs-node1"},
			false},
		{"parse bootstrap",
			args{"1", "ipfs-node1", map[string]string{keyBootstrapPeers: string(b)}},
			NodeInfo{DockerID: "1", ContainerName: "ipfs-node1", BootstrapPeers: bl},
			false},
		{"parse ports",
			args{"1", "ipfs-node1", map[string]string{keyPortAPI: "8080-8090"}},
			NodeInfo{DockerID: "1", ContainerName: "ipfs-node1", Ports: NodePorts{
				API: "8080-8090",
			}},
			false},
		{"parse resources",
			args{"1", "ipfs-node1", map[string]string{keyResourcesMemory: "4"}},
			NodeInfo{DockerID: "1", ContainerName: "ipfs-node1", Resources: NodeResources{
				MemoryGB: 4,
			}},
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

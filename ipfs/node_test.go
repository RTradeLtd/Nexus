package ipfs

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
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

func TestNodeInfo_updateFromContainerDetails(t *testing.T) {
	type args struct {
		c *types.Container
	}
	tests := []struct {
		name string
		args args
		want NodeInfo
	}{
		{"nil container", args{nil}, NodeInfo{}},
		{"with container", args{&types.Container{
			ID: "abcde",
			Ports: []types.Port{
				{PrivatePort: 4001, PublicPort: 3456},
				{PrivatePort: 5001, PublicPort: 2345},
				{PrivatePort: 8080, PublicPort: 1234},
			},
		}}, NodeInfo{
			DockerID: "abcde",
			Ports: NodePorts{
				Swarm:   "3456",
				API:     "2345",
				Gateway: "1234",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n = NodeInfo{}
			n.updateFromContainerDetails(tt.args.c)
			if !reflect.DeepEqual(n, tt.want) {
				t.Errorf("updateFromContainerDetails() = %v, want %v", n, tt.want)
			}
		})
	}
}

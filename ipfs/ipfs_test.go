package ipfs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

const defaultIPFSVersion = "v0.4.17"

func testClient() (*client, error) {
	ipfsImage := "ipfs/go-ipfs:" + defaultIPFSVersion
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	_, err = d.ImagePull(context.Background(), ipfsImage, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download IPFS image: %s", err.Error())
	}

	return &client{ipfsImage, d}, nil
}

func TestNewClient(t *testing.T) {
	_, err := NewClient(config.IPFS{Version: defaultIPFSVersion})
	if err != nil {
		t.Error(err)
	}
}

func Test_client_CreateNode(t *testing.T) {
	c, err := testClient()
	if err != nil {
		t.Error(err)
		return
	}
	key, err := SwarmKey()
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		n    *NodeInfo
		opts NodeOpts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"invalid config", args{
			&NodeInfo{"test1", NodePorts{"4001", "5001", "8080"}, ""},
			NodeOpts{},
		}, true},
		{"new node", args{
			&NodeInfo{"test1", NodePorts{"4001", "5001", "8080"}, ""},
			NodeOpts{[]byte(key), nil},
		}, true},
		{"with bootstrap", args{
			&NodeInfo{"test1", NodePorts{"4001", "5001", "8080"}, ""},
			NodeOpts{[]byte(key), []string{"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64"}},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.CreateNode(context.Background(), tt.args.n, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("client.CreateNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			time.Sleep(10 * time.Second)
			timeout := time.Duration(10 * time.Second)
			c.d.ContainerStop(context.Background(), tt.args.n.DockerID(), &timeout)
		})
	}
}

package ipfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

func init() {
	pwd, _ := os.Getwd()
	tmp := filepath.Join(pwd, "tmp")
	os.Setenv(dirEnv, tmp)
}

func testClient() (*client, error) {
	ipfsImage := "ipfs/go-ipfs:" + config.DefaultIPFSVersion
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	_, err = d.ImagePull(context.Background(), ipfsImage, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download IPFS image: %s", err.Error())
	}

	l, _ := log.NewTestLogger()
	return &client{l, d, ipfsImage}, nil
}

func TestNewClient(t *testing.T) {
	l, _ := log.NewTestLogger()
	_, err := NewClient(l, config.IPFS{Version: config.DefaultIPFSVersion})
	if err != nil {
		t.Error(err)
	}
}

func Test_client_CreateNode_GetNode(t *testing.T) {
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

	// test watcher
	eventCount := 0
	shouldGetEvents := 0
	watchCtx, cancelWatch := context.WithCancel(context.Background())
	go func() {
		events, errs := c.Watch(watchCtx)
		for {
			select {
			case err := <-errs:
				if err != nil {
					t.Log(err.Error())
				}
			case event := <-events:
				eventCount++
				t.Logf("event received: %v\n", event)
			}
		}
	}()

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
			&NodeInfo{"test1", NodePorts{"4001", "5001", "8080"}, "", "", "", nil},
			NodeOpts{},
		}, true},
		{"new node", args{
			&NodeInfo{"test2", NodePorts{"4001", "5001", "8080"}, "", "", "", nil},
			NodeOpts{[]byte(key), nil, true},
		}, false},
		{"with bootstrap", args{
			&NodeInfo{"test3", NodePorts{"4001", "5001", "8080"}, "", "", "", nil},
			NodeOpts{[]byte(key),
				[]string{
					"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
					"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
				},
				true},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// create node
			if err := c.CreateNode(ctx, tt.args.n, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("client.CreateNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// check that container is up, watcher should receive an event
			shouldGetEvents++
			time.Sleep(1 * time.Second)
			n, err := c.Nodes(ctx)
			if err != nil {
				t.Error(err.Error())
				return
			}
			found := false
			for _, node := range n {
				if node.DockerID() == tt.args.n.DockerID() {
					found = true
				}
			}
			if !found {
				t.Errorf("could not find container %s", tt.args.n.DockerID())
			}

			// clean up, watcher should receive an event
			c.StopNode(ctx, tt.args.n)
			shouldGetEvents++
		})
	}

	cancelWatch()
	if shouldGetEvents != eventCount {
		t.Errorf("expected %d events, got %d", shouldGetEvents, eventCount)
	}
}

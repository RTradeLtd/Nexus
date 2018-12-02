package ipfs

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	docker "github.com/docker/docker/client"
)

func testClient() (*client, error) {
	ipfsImage := "ipfs/go-ipfs:" + config.DefaultIPFSVersion
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	l, _ := log.NewLogger("", true)
	return &client{l, d, ipfsImage, "./tmp", 0755}, nil
}

func TestNewClient(t *testing.T) {
	l, _ := log.NewTestLogger()
	_, err := NewClient(l, config.IPFS{
		Version:  config.DefaultIPFSVersion,
		ModePerm: "0700",
	})
	if err != nil {
		t.Error(err)
	}
}

func Test_client_NodeOperations(t *testing.T) {
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
	var eventCount int
	var shouldGetEvents int
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
			&NodeInfo{"test1", "", NodePorts{"4001", "5001", "8080"}, NodeResources{}, "", "", "", nil},
			NodeOpts{},
		}, true},
		{"new node", args{
			&NodeInfo{"test2", "", NodePorts{"4001", "5001", "8080"}, NodeResources{}, "", "", "", nil},
			NodeOpts{[]byte(key), nil, false},
		}, false},
		{"with bootstrap", args{
			&NodeInfo{"test3", "", NodePorts{"4001", "5001", "8080"}, NodeResources{}, "", "", "", nil},
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

			// clean up afterwards
			defer func() {
				c.StopNode(ctx, tt.args.n)
				c.RemoveNode(ctx, tt.args.n.NetworkID)
			}()

			// check that container is up, watcher should receive an event
			shouldGetEvents++
			time.Sleep(1 * time.Second)
			n, err := c.Nodes(ctx)
			if err != nil {
				t.Error(err.Error())
				return
			}
			for _, node := range n {
				if node.DockerID == tt.args.n.DockerID {
					goto FOUND
				}
			}
			t.Errorf("could not find container %s", tt.args.n.DockerID)
			return

		FOUND:
			// should receive a cleanup event
			shouldGetEvents++

			// get node stats
			s, err := c.NodeStats(ctx, tt.args.n)
			if err != nil {
				t.Error(err.Error())
				return
			}
			t.Logf("received stats: %v", s)

			// stop node
			c.StopNode(ctx, tt.args.n)
		})
	}

	cancelWatch()
	if shouldGetEvents != eventCount {
		t.Errorf("expected %d events, got %d", shouldGetEvents, eventCount)
	}
}

func Test_client_getDataDir(t *testing.T) {
	c, err := testClient()
	if err != nil {
		t.Error(err)
		return
	}
	d := c.getDataDir("path")
	if !strings.Contains(d, "path") {
		t.Errorf("expected 'path' in path, got %s", d)
	}
	if !strings.HasPrefix(d, "/") {
		t.Errorf("expected absolute path, got %s", d)
	}
}

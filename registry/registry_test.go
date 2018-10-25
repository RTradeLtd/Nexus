package registry

import (
	"fmt"
	"net"
	"testing"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
)

var defaultNode = ipfs.NodeInfo{
	Network: "bobheadxi",
	Ports:   ipfs.NodePorts{Swarm: "4001", API: "5001", Gateway: "8080"},
}

func newTestRegistry() *NodeRegistry {
	// create a registry with a mock node for testing
	n := defaultNode
	return New(config.New().Ports, &n)
}

func TestNew(t *testing.T) {
	r := newTestRegistry()
	r.Close()
}

func TestNodeRegistry_Register(t *testing.T) {
	r := newTestRegistry()
	defer r.Close()

	cfg := config.New().Ports
	cfg.API = []string{}
	fmt.Printf("%v\n", cfg)
	rNoSwarm := New(cfg)
	defer rNoSwarm.Close()

	cfg = config.New().Ports
	cfg.API = []string{}
	rNoAPI := New(cfg)
	defer rNoAPI.Close()

	cfg = config.New().Ports
	cfg.API = []string{}
	rNoGateway := New(cfg)
	defer rNoGateway.Close()

	type args struct {
		node *ipfs.NodeInfo
	}
	tests := []struct {
		name    string
		reg     *NodeRegistry
		args    args
		wantErr bool
	}{
		{"invalid input", r, args{&ipfs.NodeInfo{}}, true},
		{"existing network", r, args{&ipfs.NodeInfo{Network: "bobheadxi"}}, true},
		{"no swarm port", rNoSwarm, args{&ipfs.NodeInfo{Network: "timhortons"}}, true},
		{"no api port", rNoAPI, args{&ipfs.NodeInfo{Network: "kfc"}}, true},
		{"no gateway port", rNoGateway, args{&ipfs.NodeInfo{Network: "mcdonalds"}}, true},
		{"successful registration", r, args{&ipfs.NodeInfo{Network: "postables"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.reg.Register(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistry.Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			// check if port assignment should be empty
			if tt.wantErr && tt.args.node.Ports.Swarm != "" {
				t.Error("port should be unassigned")
			} else if !tt.wantErr && tt.args.node.Ports.Swarm == "" {
				t.Error("port should be assigned")
			}

			if tt.args.node.Ports.Swarm != "" {
				// check if ports were released or still locked
				if _, err := net.Listen("tcp", "0.0.0.0:"+tt.args.node.Ports.Swarm); (err != nil) != tt.wantErr {
					t.Errorf("expected port %s locked=%v", tt.args.node.Ports.Swarm, tt.wantErr)
				}
				if _, err := net.Listen("tcp", "127.0.0.1:"+tt.args.node.Ports.API); (err != nil) != tt.wantErr {
					t.Errorf("expected port %s locked=%v", tt.args.node.Ports.API, tt.wantErr)
				}
				if _, err := net.Listen("tcp", "127.0.0.1:"+tt.args.node.Ports.Gateway); (err != nil) != tt.wantErr {
					t.Errorf("expected port %s locked=%v", tt.args.node.Ports.API, tt.wantErr)
				}
			}

		})
	}
}

func TestNodeRegistry_Deregister(t *testing.T) {
	type args struct {
		network string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"invalid input", args{""}, true},
		{"unknown network", args{"timhortons"}, true},
		{"successful deregistration", args{"bobheadxi"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRegistry()
			defer r.Close()
			if err := r.Deregister(tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistry.Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// check ports to default node are locked
				if _, err := net.Listen("tcp", "0.0.0.0:"+defaultNode.Ports.Swarm); err == nil {
					t.Errorf("port %s shoud be blocked", defaultNode.Ports.Swarm)
				}
				if _, err := net.Listen("tcp", "127.0.0.1:"+defaultNode.Ports.API); err == nil {
					t.Errorf("port %s shoud be blocked", defaultNode.Ports.API)
				}
				if _, err := net.Listen("tcp", "127.0.0.1:"+defaultNode.Ports.Gateway); err == nil {
					t.Errorf("port %s shoud be blocked", defaultNode.Ports.Gateway)
				}
			}
		})
	}
}

func TestNodeRegistry_List(t *testing.T) {
	r := newTestRegistry()
	defer r.Close()
	nodes := r.List()
	if len(nodes) != len(r.nodes) {
		t.Errorf("expected %d nodes, got %d", len(nodes), len(r.nodes))
	}
}

func TestNodeRegistry_Get(t *testing.T) {
	r := newTestRegistry()
	defer r.Close()
	type args struct {
		network string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"invalid input", args{""}, true},
		{"unknown node", args{"maccas"}, true},
		{"valid node", args{"bobheadxi"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Get(tt.args.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistry.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

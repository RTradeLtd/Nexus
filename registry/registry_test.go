package registry

import (
	"net"
	"testing"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
)

func newTestRegistry() *NodeRegistry {
	// create a registry with a mock node for testing
	return New(config.New().Ports, &ipfs.NodeInfo{
		Network: "bobheadxi",
		Ports:   ipfs.NodePorts{Swarm: "4001", API: "5001", Gateway: "8080"},
	})
}

func TestNew(t *testing.T) {
	newTestRegistry()
}

func TestNodeRegistry_Register(t *testing.T) {
	r := newTestRegistry()

	cfg := config.New().Ports
	cfg.API = []string{}
	rNoSwarm := New(cfg)

	cfg = config.New().Ports
	cfg.API = []string{}
	rNoAPI := New(cfg)

	cfg = config.New().Ports
	cfg.API = []string{}
	rNoGateway := New(cfg)

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
		{"no swarm port", rNoSwarm, args{&ipfs.NodeInfo{Network: "bobheadxi"}}, true},
		{"no api port", rNoAPI, args{&ipfs.NodeInfo{Network: "bobheadxi"}}, true},
		{"no gateway port", rNoGateway, args{&ipfs.NodeInfo{Network: "bobheadxi"}}, true},
		{"successful registration", r, args{&ipfs.NodeInfo{Network: "postables"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.Register(tt.args.node); (err != nil) != tt.wantErr {
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
				if _, err := net.Listen("tcp", "127.0.0.1:"+tt.args.node.Ports.Swarm); (err != nil) != tt.wantErr {
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

package registry

import (
	"testing"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
)

var defaultNode = ipfs.NodeInfo{
	NetworkID: "bobheadxi",
	Ports:     ipfs.NodePorts{Swarm: "4001", API: "5001", Gateway: "8080"},
}

func newTestRegistry() *NodeRegistry {
	// create a registry with a mock node for testing
	n := defaultNode
	l, _ := log.NewTestLogger()
	return New(l, config.New().Ports, &n)
}

func TestNew(t *testing.T) {
	newTestRegistry()
}

func TestNodeRegistry_Register(t *testing.T) {
	r := newTestRegistry()
	l, _ := log.NewTestLogger()

	cfg := config.New().Ports
	cfg.Swarm = []string{}
	rNoSwarm := New(l, cfg)

	cfg = config.New().Ports
	cfg.API = []string{}
	rNoAPI := New(l, cfg)

	cfg = config.New().Ports
	cfg.Gateway = []string{}
	rNoGateway := New(l, cfg)

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
		{"existing network", r, args{&ipfs.NodeInfo{NetworkID: "bobheadxi"}}, true},
		{"no swarm port", rNoSwarm, args{&ipfs.NodeInfo{NetworkID: "timhortons"}}, true},
		{"no api port", rNoAPI, args{&ipfs.NodeInfo{NetworkID: "kfc"}}, true},
		{"no gateway port", rNoGateway, args{&ipfs.NodeInfo{NetworkID: "mcdonalds"}}, true},
		{"successful registration", r, args{&ipfs.NodeInfo{NetworkID: "postables"}}, false},
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
			if err := r.Deregister(tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistry.Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistry_List(t *testing.T) {
	r := newTestRegistry()
	nodes := r.List()
	if len(nodes) != len(r.nodes) {
		t.Errorf("expected %d nodes, got %d", len(nodes), len(r.nodes))
	}
}

func TestNodeRegistry_Get(t *testing.T) {
	r := newTestRegistry()
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

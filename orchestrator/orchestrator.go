package orchestrator

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
)

// Orchestrator contains most primary application logic and manages node
// availability
type Orchestrator struct {
	client ipfs.NodeClient
	reg    *registry.NodeRegistry
}

// New instantiates and bootstraps a new Orchestrator
func New(ipfsOpts config.IPFS, pgOpts config.Postgres) (*Orchestrator, error) {
	c, err := ipfs.NewClient(ipfsOpts)
	if err != nil {
		return nil, err
	}

	// bootstrap registry
	nodes, err := c.Nodes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch nodes: %s", err.Error())
	}
	reg := registry.New(nodes...)

	return &Orchestrator{client: c, reg: reg}, nil
}

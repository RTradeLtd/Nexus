package orchestrator

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"go.uber.org/zap"
)

// Orchestrator contains most primary application logic and manages node
// availability
type Orchestrator struct {
	l *zap.SugaredLogger

	client ipfs.NodeClient
	reg    *registry.NodeRegistry
}

// New instantiates and bootstraps a new Orchestrator
func New(logger *zap.SugaredLogger, ipfsOpts config.IPFS, pgOpts config.Postgres) (*Orchestrator, error) {
	c, err := ipfs.NewClient(logger, ipfsOpts)
	if err != nil {
		return nil, err
	}

	// bootstrap registry
	nodes, err := c.Nodes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch nodes: %s", err.Error())
	}
	reg := registry.New(logger, ipfsOpts.Ports, nodes...)

	return &Orchestrator{l: logger, client: c, reg: reg}, nil
}

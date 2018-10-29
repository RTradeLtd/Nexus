package orchestrator

import (
	"context"
	"fmt"

	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"go.uber.org/zap"
)

// Orchestrator contains most primary application logic and manages node
// availability
type Orchestrator struct {
	l   *zap.SugaredLogger
	dbm *database.DatabaseManager

	client ipfs.NodeClient
	reg    *registry.NodeRegistry
}

// New instantiates and bootstraps a new Orchestrator
func New(logger *zap.SugaredLogger, c ipfs.NodeClient, ports config.Ports, pg tcfg.Database, dev bool) (*Orchestrator, error) {
	// bootstrap registry
	nodes, err := c.Nodes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch nodes: %s", err.Error())
	}
	reg := registry.New(logger, ports, nodes...)

	// set up database connection
	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: pg,
	}, database.DatabaseOptions{
		SSLModeDisable: dev,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s", err.Error())
	}

	return &Orchestrator{l: logger, dbm: dbm, client: c, reg: reg}, nil
}

// Run initalizes the orchestrator's background tasks
func (o *Orchestrator) Run(ctx context.Context) error {
	go o.client.Watch(ctx)
	go func() {
		select {
		case <-ctx.Done():
			if err := o.dbm.Close(); err != nil {
				o.l.Warnf("error occured closing database connection",
					"error", err)
			}
		}
	}()
	return nil
}

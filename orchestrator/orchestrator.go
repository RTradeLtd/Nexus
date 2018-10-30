package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"go.uber.org/zap"
)

// Orchestrator contains most primary application logic and manages node
// availability
type Orchestrator struct {
	l  *zap.SugaredLogger
	nm *models.IPFSNetworkManager

	client ipfs.NodeClient
	reg    *registry.NodeRegistry
	host   string
}

// New instantiates and bootstraps a new Orchestrator
func New(logger *zap.SugaredLogger, host string, c ipfs.NodeClient,
	ports config.Ports, pg tcfg.Database, dev bool) (*Orchestrator, error) {
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

	return &Orchestrator{
		l:      logger,
		nm:     models.NewHostedIPFSNetworkManager(dbm.DB),
		client: c,
		reg:    reg,
	}, nil
}

// Run initalizes the orchestrator's background tasks
func (o *Orchestrator) Run(ctx context.Context) error {
	go o.client.Watch(ctx)
	go func() {
		select {
		case <-ctx.Done():
			if err := o.nm.DB.Close(); err != nil {
				o.l.Warnw("error occured closing database connection",
					"error", err)
			}
		}
	}()
	return nil
}

// NetworkUp intializes a node for given network
func (o *Orchestrator) NetworkUp(ctx context.Context, network string) error {
	if network == "" {
		return errors.New("invalid network name provided")
	}

	start := time.Now()
	id := generateID()
	l := log.NewProcessLogger(o.l, "network_up",
		"id", id,
		"network", network)
	l.Info("network up process started")

	// check if request is valid
	n, err := o.nm.GetNetworkByName(network)
	if err != nil {
		return fmt.Errorf("no network with name '%s' found", network)
	}

	// set options based on database entry
	opts, err := getOptionsFromDatabaseEntry(n)
	if err != nil {
		return fmt.Errorf("failed to configure network: %s", err.Error())
	}

	// register node for network
	newNode := &ipfs.NodeInfo{Network: network}
	if err := o.reg.Register(newNode); err != nil {
		return fmt.Errorf("failed to allocate resources for network '%s': %s", network, err)
	}

	// instantiate node
	l.Info("creating node")
	if err := o.client.CreateNode(ctx, newNode, opts); err != nil {
		return fmt.Errorf("failed to instantiate node for network '%s': %s", network, err)
	}
	l.Infow("node created",
		"node.docker_id", newNode.DockerID(),
		"node.data_dir", newNode.DataDirectory(),
		"node.ports", newNode.Ports)

	// update network in database
	n.APIURL = o.host + newNode.Ports.API
	n.SwarmKey = string(opts.SwarmKey)
	n.Activated = time.Now()
	if err := o.nm.UpdateNetwork(n); err != nil {
		return fmt.Errorf("failed to update network '%s': %s", network, err)
	}

	l.Infow("network up process completed",
		"network_up.duration", time.Since(start))

	return nil
}

package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/network"
	"go.uber.org/zap"
)

const (
	// ErrInvalidNetwork is returned when an invalid node ID is provided
	ErrInvalidNetwork = "invalid node network"

	// ErrNetworkExists is returned when an network is provided that already exists
	ErrNetworkExists = "network already exists"
)

// NodeRegistry manages data on active nodes
type NodeRegistry struct {
	l *zap.SugaredLogger

	// node registry - locked by NodeRegistry::nm
	nodes map[string]*ipfs.NodeInfo
	nm    sync.RWMutex

	// port registry
	swarmPorts   *network.Registry
	apiPorts     *network.Registry
	gatewayPorts *network.Registry
}

// New sets up a new registry with provided nodes
func New(logger *zap.SugaredLogger, ports config.Ports, nodes ...*ipfs.NodeInfo) *NodeRegistry {
	// parse nodes
	m := make(map[string]*ipfs.NodeInfo)
	if nodes != nil {
		for _, n := range nodes {
			m[n.Network] = n
		}
	}

	// build registry
	return &NodeRegistry{
		l:     logger.Named("registry"),
		nodes: m,

		swarmPorts:   network.NewRegistry(logger, "0.0.0.0", ports.Swarm),
		apiPorts:     network.NewRegistry(logger, "127.0.0.1", ports.API),
		gatewayPorts: network.NewRegistry(logger, "127.0.0.1", ports.Gateway),
	}
}

// Register registers a node and allocates appropriate ports
func (r *NodeRegistry) Register(node *ipfs.NodeInfo) error {
	if node.Network == "" {
		return errors.New(ErrInvalidNetwork)
	}

	r.nm.Lock()
	defer r.nm.Unlock()

	if _, found := r.nodes[node.Network]; found {
		return errors.New(ErrNetworkExists)
	}

	// assign ports to this node
	var err error
	var swarm, api, gateway string
	if swarm, err = r.swarmPorts.AssignPort(); err != nil {
		return fmt.Errorf("failed to register node: %s", err.Error())
	}
	if api, err = r.apiPorts.AssignPort(); err != nil {
		return fmt.Errorf("failed to register node: %s", err.Error())
	}
	if gateway, err = r.gatewayPorts.AssignPort(); err != nil {
		return fmt.Errorf("failed to register node: %s", err.Error())
	}
	node.Ports = ipfs.NodePorts{Swarm: swarm, API: api, Gateway: gateway}

	r.nodes[node.Network] = node

	return nil
}

// Deregister removes node with given network
func (r *NodeRegistry) Deregister(network string) error {
	if network == "" {
		return errors.New(ErrInvalidNetwork)
	}

	r.nm.Lock()
	defer r.nm.Unlock()

	if _, found := r.nodes[network]; !found {
		return fmt.Errorf("node for network '%s' not found", network)
	}

	delete(r.nodes, network)
	return nil
}

// List retrieves a list of all known nodes
func (r *NodeRegistry) List() []ipfs.NodeInfo {
	var (
		nodes = make([]ipfs.NodeInfo, len(r.nodes))
		i     = 0
	)

	r.nm.RLock()
	for _, n := range r.nodes {
		nodes[i] = *n
		i++
	}
	r.nm.RUnlock()

	return nodes
}

// Get retrieves details about node with given network
func (r *NodeRegistry) Get(network string) (ipfs.NodeInfo, error) {
	var node ipfs.NodeInfo
	if network == "" {
		return node, errors.New(ErrInvalidNetwork)
	}

	r.nm.RLock()
	n, found := r.nodes[network]
	if !found {
		r.nm.RUnlock()
		return node, fmt.Errorf("node for network '%s' not found", network)
	}
	node = *n
	r.nm.RUnlock()

	return node, nil
}

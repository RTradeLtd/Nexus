package registry

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
)

const (
	// ErrInvalidNetwork is returned when an invalid node ID is provided
	ErrInvalidNetwork = "invalid node network"

	// ErrNetworkExists is returned when an network is provided that already exists
	ErrNetworkExists = "network already exists"
)

// NodeRegistry manages data on active nodes
type NodeRegistry struct {
	// node registry - locked by NodeRegistry::nm
	nodes map[string]*ipfs.NodeInfo
	nm    sync.RWMutex

	// port registry - locked by NodeRegistry::pm
	swarmPorts   map[string]net.Listener
	apiPorts     map[string]net.Listener
	gatewayPorts map[string]net.Listener
	pm           sync.RWMutex
}

// New sets up a new registry with provided nodes
func New(ports config.Ports, nodes ...*ipfs.NodeInfo) *NodeRegistry {
	var (
		m       = make(map[string]*ipfs.NodeInfo)
		swarm   = make(map[string]net.Listener)
		api     = make(map[string]net.Listener)
		gateway = make(map[string]net.Listener)
	)

	// parse nodes
	if nodes != nil {
		for _, n := range nodes {
			m[n.Network] = n
		}
	}

	// parse all port ranges and register them, locking with net listeners if they
	// are available
	if ports.Swarm != nil {
		pts := parsePorts(ports.Swarm)
		for _, p := range pts {
			swarm[p], _ = net.Listen("tcp", "0.0.0.0:"+p)
		}
	}
	if ports.API != nil {
		pts := parsePorts(ports.API)
		for _, p := range pts {
			api[p], _ = net.Listen("tcp", ":"+p)
		}
	}
	if ports.Gateway != nil {
		pts := parsePorts(ports.Gateway)
		for _, p := range pts {
			gateway[p], _ = net.Listen("tcp", ":"+p)
		}
	}

	// build registry
	return &NodeRegistry{
		nodes:        m,
		swarmPorts:   swarm,
		apiPorts:     api,
		gatewayPorts: gateway,
	}
}

// Register registers a node and allocates appropriate ports
func (r *NodeRegistry) Register(node *ipfs.NodeInfo) error {
	if node.DockerID() == "" {
		return errors.New(ErrInvalidNetwork)
	}
	r.nm.Lock()
	if _, found := r.nodes[node.Network]; found {
		return errors.New(ErrNetworkExists)
	}
	r.nodes[node.DockerID()] = node
	r.nm.Unlock()
	return nil
}

// Deregister removes node with given network
func (r *NodeRegistry) Deregister(network string) error {
	if network == "" {
		return errors.New(ErrInvalidNetwork)
	}
	r.nm.Lock()
	if _, found := r.nodes[network]; !found {
		return fmt.Errorf("node for network '%s' not found", network)
	}
	delete(r.nodes, network)
	r.nm.Unlock()
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
		return node, fmt.Errorf("node for network '%s' not found", network)
	}
	node = *n
	r.nm.RUnlock()
	return node, nil
}

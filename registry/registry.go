package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
)

const (
	ErrInvalidID = "invalid node ID"
)

type NodeRegistry struct {
	nodes map[string]*ipfs.NodeInfo
	m     sync.RWMutex
}

func New(nodes ...*ipfs.NodeInfo) *NodeRegistry {
	m := make(map[string]*ipfs.NodeInfo)
	if nodes != nil {
		for _, n := range nodes {
			m[n.DockerID()] = n
		}
	}
	return &NodeRegistry{nodes: m}
}

func (r *NodeRegistry) Register(node ipfs.NodeInfo) error {
	if node.DockerID() == "" {
		return errors.New(ErrInvalidID)
	}
	r.m.Lock()
	if _, found := r.nodes[node.DockerID()]; found {
		return errors.New("node ID already exists")
	}
	r.nodes[node.DockerID()] = &node
	r.m.Unlock()
	return nil
}

func (r *NodeRegistry) Deregister(id string) error {
	if id == "" {
		return errors.New(ErrInvalidID)
	}
	r.m.Lock()
	if _, found := r.nodes[id]; !found {
		return fmt.Errorf("node with ID '%s' not found", id)
	}
	delete(r.nodes, id)
	r.m.Unlock()
	return nil
}

func (r *NodeRegistry) List() []ipfs.NodeInfo {
	var (
		nodes = make([]ipfs.NodeInfo, len(r.nodes))
		i     = 0
	)

	r.m.RLock()
	for _, n := range r.nodes {
		nodes[i] = *n
	}
	r.m.RUnlock()

	return nodes
}

func (r *NodeRegistry) Get(id string) (ipfs.NodeInfo, error) {
	var node ipfs.NodeInfo
	if id == "" {
		return node, errors.New(ErrInvalidID)
	}
	r.m.RLock()
	n, found := r.nodes[id]
	if !found {
		return node, fmt.Errorf("node with ID '%s' not found", id)
	}
	node = *n
	r.m.RUnlock()
	return node, nil
}

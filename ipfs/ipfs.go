package ipfs

import (
	"context"

	docker "github.com/docker/docker/client"
)

// NodeClient provides an interface to the base Docker client for controlling
// IPFS nodes
type NodeClient interface {
	// Nodes retrieves a list of active IPFS ndoes
	Nodes() ([]*NodeInfo, error)

	// CreateNode activates a new IPFS node
	CreateNode(n NodeInfo) error

	// StopNode shuts down an existing IPFS node
	StopNode(id string) error
}

type client struct {
	c *docker.Client
}

// NewClient creates a new Docker Client from ENV values and negotiates the
// correct API version to use
func NewClient() (NodeClient, error) {
	c, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}
	c.NegotiateAPIVersion(context.Background())
	return &client{c}, nil
}

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	ID   string
	Port string
}

func (c *client) Nodes() ([]*NodeInfo, error) {
	return nil, nil
}

func (c *client) CreateNode(n NodeInfo) error {
	return nil
}

func (c *client) StopNode(id string) error {
	return nil
}

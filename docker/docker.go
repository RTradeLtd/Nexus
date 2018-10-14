package docker

import (
	"context"

	docker "github.com/docker/docker/client"
)

// Client provides an interface to the base Docker client
type Client interface {
}

type client struct {
	c *docker.Client
}

// NewClient creates a new Docker Client from ENV values and negotiates the
// correct API version to use
func NewClient() (Client, error) {
	c, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}
	c.NegotiateAPIVersion(context.Background())
	return &client{c}, nil
}

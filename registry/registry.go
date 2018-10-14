package registry

import "github.com/RTradeLtd/ipfs-orchestrator/docker"

type Manager struct {
	docker.Client
}

func New() (*Manager, error) {
	c, err := docker.NewClient()
	if err != nil {
		return nil, err
	}
	return &Manager{c}, nil
}

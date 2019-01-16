package ipfs

import (
	"context"
	"fmt"
	"testing"

	"github.com/RTradeLtd/Nexus/config"
	"github.com/RTradeLtd/Nexus/log"
	docker "github.com/docker/docker/client"
)

func newTestClient() (NodeClient, error) {
	ipfsImage := "ipfs/go-ipfs:" + config.DefaultIPFSVersion
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	l, _ := log.NewLogger("", true)
	return &Client{l, d, ipfsImage, "./tmp", 0755}, nil
}

func TestNewClient(t *testing.T) {
	l, _ := log.NewTestLogger()
	_, err := NewClient(l, config.IPFS{
		Version:  config.DefaultIPFSVersion,
		ModePerm: "0700",
	})
	if err != nil {
		t.Error(err)
	}
}

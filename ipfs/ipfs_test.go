package ipfs

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/RTradeLtd/Nexus/config"
	"github.com/RTradeLtd/Nexus/log"
	docker "github.com/docker/docker/client"
)

func newTestClient(t *testing.T) (NodeClient, error) {
	ipfsImage := "ipfs/go-ipfs:" + config.DefaultIPFSVersion
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	var l = zaptest.NewLogger(t).Sugar()
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

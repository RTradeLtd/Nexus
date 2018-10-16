package ipfs

import (
	"testing"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient(config.IPFS{Version: "v0.4.17"})
	if err != nil {
		t.Error(err)
	}
}

package ipfs

import (
	"strings"
	"testing"
)

func TestSwarmKey(t *testing.T) {
	key, err := SwarmKey()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(key, "/key/swarm/psk/1.0.0/") {
		t.Error("key signature not found")
	}
}

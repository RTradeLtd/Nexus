package ipfs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// SwarmKey generates a new swarm key
func SwarmKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("error generating key: %s", err.Error())
	}
	return "/key/swarm/psk/1.0.0/\n/base16/\n" + hex.EncodeToString(key), nil
}

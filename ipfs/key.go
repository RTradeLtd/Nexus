package ipfs

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

// SwarmKey generates a new swarm key
func SwarmKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalln("While trying to read random source:", err)
	}
	return "/key/swarm/psk/1.0.0/\n/base16/\n" + hex.EncodeToString(key), nil
}

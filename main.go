package main

import (
	"log"

	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
)

func main() {
	r, err := registry.New()
	if err != nil {
		log.Fatal(err)
	}

	d, err := daemon.New(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(d.Run("localhost", "9111"))
}

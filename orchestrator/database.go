package orchestrator

import (
	"errors"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
)

func getOptionsFromDatabaseEntry(network *models.HostedIPFSPrivateNetwork) (ipfs.NodeOpts, error) {
	opts := ipfs.NodeOpts{}
	if network == nil {
		return opts, errors.New("invalid network entry")
	}

	// set swarm key
	if network.SwarmKey != "" {
		opts.SwarmKey = []byte(network.SwarmKey)
	} else {
		key, err := ipfs.SwarmKey()
		if err != nil {
			return opts, err
		}
		opts.SwarmKey = []byte(key)
	}

	// set bootstrap ppers
	opts.BootstrapPeers = network.BootstrapPeerAddresses

	return opts, nil
}

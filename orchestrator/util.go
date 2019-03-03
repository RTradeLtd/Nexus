package orchestrator

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func generateID() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}

func rebootOfflineNodes(orch *Orchestrator) {
	offlineNetworks, err := orch.nm.GetOfflineNetworks(false)
	if err != nil {
		orch.l.Errorw("unabled to fetch offline nodes", "error", err)
	} else if len(offlineNetworks) > 0 {
		for _, network := range offlineNetworks {
			if network.Disabled {
				continue
			}
			if _, err := orch.Registry.Get(network.Name); err != nil {
				var nl = orch.l.With("network", network.Name)
				nl.Infow("rebooting network",
					"network", network.Name,
					"network.db_id", network.ID)
				d, err := orch.NetworkUp(context.Background(), network.Name)
				if err != nil {
					nl.Errorw("failed to reboot network", "network", network.Name)
				} else {
					nl.Infow("node started", "network.details", d)
				}
			}
		}
	}
}

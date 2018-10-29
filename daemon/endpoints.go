package daemon

import (
	"context"

	ipfs_orchestrator "github.com/RTradeLtd/ipfs-orchestrator/protobuf"
)

// Ping is useful for checking client-server connection
func (d *Daemon) Ping(c context.Context, req *ipfs_orchestrator.Empty) (*ipfs_orchestrator.Empty, error) {
	return &ipfs_orchestrator.Empty{}, nil
}

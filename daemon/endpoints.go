package daemon

import (
	"context"

	orchestrator "github.com/RTradeLtd/ipfs-orchestrator/protobuf"
)

// Ping is useful for checking client-server connection
func (d *Daemon) Ping(c context.Context, req *orchestrator.Empty) (*orchestrator.Empty, error) {
	return &orchestrator.Empty{}, nil
}

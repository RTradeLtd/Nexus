package daemon

import (
	"context"

	ipfs_orchestrator "github.com/RTradeLtd/ipfs-orchestrator/protobuf"
)

// Ping is useful for checking client-server connection
func (d *Daemon) Ping(c context.Context,
	req *ipfs_orchestrator.Empty) (*ipfs_orchestrator.Empty, error) {
	return &ipfs_orchestrator.Empty{}, nil
}

// StartNetwork brings a node for the requested network online
func (d *Daemon) StartNetwork(ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest) (*ipfs_orchestrator.StartNetworkResponse, error) {

	n, err := d.o.NetworkUp(ctx, req.GetNetwork())
	if err != nil {
		return nil, err
	}

	return &ipfs_orchestrator.StartNetworkResponse{
		Api:      n.API,
		SwarmKey: n.SwarmKey,
	}, nil
}

// StopNetwork brings a node for the requested network offline
func (d *Daemon) StopNetwork(ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest) (*ipfs_orchestrator.Empty, error) {

	return &ipfs_orchestrator.Empty{}, d.o.NetworkDown(ctx, req.GetNetwork())
}

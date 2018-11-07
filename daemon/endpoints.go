package daemon

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

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

// NetworkStats retrieves stats about the requested node
func (d *Daemon) NetworkStats(ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest) (*ipfs_orchestrator.NetworkStatusReponse, error) {

	s, err := d.o.NetworkStatus(ctx, req.Network)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	sb, err := json.Marshal(s.Stats)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &ipfs_orchestrator.NetworkStatusReponse{
		Network:   req.Network,
		Api:       s.API,
		Uptime:    int64(s.Uptime),
		DiskUsage: s.DiskUsage,
		Stats:     sb,
	}, nil
}

package daemon

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	ipfs_orchestrator "github.com/RTradeLtd/grpc/ipfs-orchestrator"
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
		SwarmKey: n.SwarmKey,
	}, nil
}

// UpdateNetwork updates the configuration of the given network
func (d *Daemon) UpdateNetwork(ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest) (*ipfs_orchestrator.Empty, error) {

	return &ipfs_orchestrator.Empty{}, d.o.NetworkUpdate(ctx, req.GetNetwork())
}

// StopNetwork brings a node for the requested network offline
func (d *Daemon) StopNetwork(ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest) (*ipfs_orchestrator.Empty, error) {

	return &ipfs_orchestrator.Empty{}, d.o.NetworkDown(ctx, req.GetNetwork())
}

// RemoveNetwork removes assets for requested node
func (d *Daemon) RemoveNetwork(ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest) (*ipfs_orchestrator.Empty, error) {

	return &ipfs_orchestrator.Empty{}, d.o.NetworkRemove(ctx, req.GetNetwork())
}

// NetworkStats retrieves stats about the requested node
func (d *Daemon) NetworkStats(
	ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest,
) (*ipfs_orchestrator.NetworkStatusReponse, error) {

	s, err := d.o.NetworkStatus(ctx, req.Network)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &ipfs_orchestrator.NetworkStatusReponse{
		Network:   req.Network,
		Uptime:    int64(s.Uptime),
		DiskUsage: s.DiskUsage,
	}, nil
}

// NetworkDiagnostics retrieves detailed diagnostic details about the requested
// network node
func (d *Daemon) NetworkDiagnostics(
	ctx context.Context,
	req *ipfs_orchestrator.NetworkRequest,
) (*ipfs_orchestrator.NetworkDiagnosticsResponse, error) {

	s, err := d.o.NetworkDiagnostics(ctx, req.Network)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	nb, err := json.Marshal(s.NodeInfo)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	sb, err := json.Marshal(s.NodeStats)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &ipfs_orchestrator.NetworkDiagnosticsResponse{
		NodeInfo: nb,
		Stats:    sb,
	}, nil
}

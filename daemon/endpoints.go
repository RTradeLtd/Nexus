package daemon

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/RTradeLtd/grpc/nexus"
)

// Ping is useful for checking client-server connection
func (d *Daemon) Ping(
	c context.Context,
	req *nexus.Empty,
) (*nexus.Empty, error) {
	return &nexus.Empty{}, nil
}

// StartNetwork brings a node for the requested network online
func (d *Daemon) StartNetwork(
	ctx context.Context,
	req *nexus.NetworkRequest,
) (*nexus.StartNetworkResponse, error) {
	n, err := d.o.NetworkUp(ctx, req.GetNetwork())
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &nexus.StartNetworkResponse{
		PeerId:    n.PeerID,
		SwarmPort: n.SwarmPort,
		SwarmKey:  n.SwarmKey,
	}, nil
}

// UpdateNetwork updates the configuration of the given network
func (d *Daemon) UpdateNetwork(
	ctx context.Context,
	req *nexus.NetworkRequest,
) (*nexus.Empty, error) {

	return &nexus.Empty{}, d.o.NetworkUpdate(ctx, req.GetNetwork())
}

// StopNetwork brings a node for the requested network offline
func (d *Daemon) StopNetwork(
	ctx context.Context,
	req *nexus.NetworkRequest,
) (*nexus.Empty, error) {

	return &nexus.Empty{}, d.o.NetworkDown(ctx, req.GetNetwork())
}

// RemoveNetwork removes assets for requested node
func (d *Daemon) RemoveNetwork(
	ctx context.Context,
	req *nexus.NetworkRequest,
) (*nexus.Empty, error) {

	return &nexus.Empty{}, d.o.NetworkRemove(ctx, req.GetNetwork())
}

// NetworkStats retrieves stats about the requested node
func (d *Daemon) NetworkStats(
	ctx context.Context,
	req *nexus.NetworkRequest,
) (*nexus.NetworkStatusReponse, error) {

	s, err := d.o.NetworkStatus(ctx, req.Network)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &nexus.NetworkStatusReponse{
		Network:   s.NetworkDetails.NetworkID,
		PeerId:    s.NetworkDetails.PeerID,
		Uptime:    int64(s.Uptime),
		DiskUsage: s.DiskUsage,
		SwarmPort: s.NetworkDetails.SwarmPort,
	}, nil
}

// NetworkDiagnostics retrieves detailed diagnostic details about the requested
// network node
func (d *Daemon) NetworkDiagnostics(
	ctx context.Context,
	req *nexus.NetworkRequest,
) (*nexus.NetworkDiagnosticsResponse, error) {

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

	return &nexus.NetworkDiagnosticsResponse{
		NodeInfo: nb,
		Stats:    sb,
	}, nil
}

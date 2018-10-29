package daemon

import (
	"context"
	"net"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
	"github.com/RTradeLtd/ipfs-orchestrator/protobuf"

	"google.golang.org/grpc"
)

// Daemon exposes orchestrator functionality via a gRPC API
type Daemon struct {
	s *grpc.Server
	o *orchestrator.Orchestrator
}

// New initializes a new Daemon
func New(o *orchestrator.Orchestrator) *Daemon {
	s := grpc.NewServer()
	d := &Daemon{
		s: s,
		o: o,
	}
	orchestrator_grpc.RegisterServiceServer(d.s, d)
	return d
}

// Run spins up daemon server
func (d *Daemon) Run(ctx context.Context, cfg config.API) error {
	listener, err := net.Listen("tcp", cfg.Host+":"+cfg.Port)
	if err != nil {
		return err
	}
	go d.o.Run(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				d.s.GracefulStop()
			}
		}
	}()
	return d.s.Serve(listener)
}

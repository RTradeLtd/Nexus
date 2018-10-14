package daemon

import (
	"net"

	"github.com/RTradeLtd/ipfs-orchestrator/registry"

	orchestrator "github.com/RTradeLtd/ipfs-orchestrator/protobuf"
	"google.golang.org/grpc"
)

type Daemon struct {
	s *grpc.Server
	r *registry.Registry
}

func New(registry *registry.Registry) *Daemon {
	s := grpc.NewServer()
	d := &Daemon{
		s: s,
		r: registry,
	}
	orchestrator.RegisterServiceServer(d.s, d)
	return d
}

func (d *Daemon) Run(host, port string) error {
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return err
	}
	return d.s.Serve(listener)
}

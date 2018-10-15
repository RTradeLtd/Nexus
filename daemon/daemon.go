package daemon

import (
	"net"

	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
	"github.com/RTradeLtd/ipfs-orchestrator/protobuf"

	"google.golang.org/grpc"
)

type Daemon struct {
	s *grpc.Server
	o *orchestrator.Orchestrator
}

func New(o *orchestrator.Orchestrator) *Daemon {
	s := grpc.NewServer()
	d := &Daemon{
		s: s,
		o: o,
	}
	orchestrator_grpc.RegisterServiceServer(d.s, d)
	return d
}

func (d *Daemon) Run(host, port string) error {
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return err
	}
	return d.s.Serve(listener)
}

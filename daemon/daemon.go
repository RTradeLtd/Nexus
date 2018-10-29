package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
	"github.com/RTradeLtd/ipfs-orchestrator/protobuf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

// Daemon exposes orchestrator functionality via a gRPC API
type Daemon struct {
	s *grpc.Server
	o *orchestrator.Orchestrator

	l *zap.SugaredLogger
}

// New initializes a new Daemon
func New(logger *zap.SugaredLogger, o *orchestrator.Orchestrator) *Daemon {
	s := grpc.NewServer()
	d := &Daemon{
		s: s,
		o: o,
		l: logger.Named("daemon"),
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

	// set logger to record all incoming requests
	grpcLogger := d.l.Desugar().Named("grpc")
	grpc_zap.ReplaceGrpcLogger(grpcLogger)
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Duration("grpc.duration", duration)
		}),
	}
	serverOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(grpcLogger, zapOpts...)),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.StreamServerInterceptor(grpcLogger, zapOpts...)),
	}

	// set up TLS if configuration provides for it
	if cfg.TLS.CertPath != "" {
		d.l.Info("setting up TLS")
		creds, err := credentials.NewServerTLSFromFile(cfg.TLS.CertPath, cfg.TLS.KeyPath)
		if err != nil {
			return fmt.Errorf("could not load TLS keys: %s", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	// start orchestrator background jobs
	go d.o.Run(ctx)

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-ctx.Done():
				d.s.GracefulStop()
			}
		}
	}()

	// spin up server
	return d.s.Serve(listener)
}

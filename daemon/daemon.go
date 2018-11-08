package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	ipfs_orchestrator "github.com/RTradeLtd/grpc/ipfs-orchestrator"
	"github.com/RTradeLtd/grpc/middleware"
	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
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
	o *orchestrator.Orchestrator
	l *zap.SugaredLogger
}

// New initializes a new Daemon
func New(logger *zap.SugaredLogger, o *orchestrator.Orchestrator) *Daemon {
	d := &Daemon{
		o: o,
		l: logger.Named("daemon"),
	}
	return d
}

// Run spins up daemon server
func (d *Daemon) Run(ctx context.Context, cfg config.API) error {
	listener, err := net.Listen("tcp", cfg.Host+":"+cfg.Port)
	if err != nil {
		return err
	}

	// set up authentication interceptor
	unaryInterceptor, streamInterceptor := middleware.NewServerInterceptors(cfg.Key)

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
			unaryInterceptor,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(grpcLogger, zapOpts...)),
		grpc_middleware.WithStreamServerChain(
			streamInterceptor,
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.StreamServerInterceptor(grpcLogger, zapOpts...)),
	}

	// set up TLS if configuration provides for it
	if cfg.TLS.CertPath != "" {
		d.l.Infow("setting up TLS",
			"cert", cfg.TLS.CertPath,
			"key", cfg.TLS.KeyPath)
		creds, err := credentials.NewServerTLSFromFile(cfg.TLS.CertPath, cfg.TLS.KeyPath)
		if err != nil {
			return fmt.Errorf("could not load TLS keys: %s", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	} else {
		d.l.Warn("no TLS configuration found")
	}

	// start orchestrator background jobs
	d.l.Info("starting orchestrator jobs")
	go d.o.Run(ctx)

	// initialize server
	server := grpc.NewServer()
	ipfs_orchestrator.RegisterServiceServer(server, d)

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-ctx.Done():
				d.l.Info("shutting down server")
				server.GracefulStop()
				return
			}
		}
	}()

	// spin up server
	d.l.Infow("spinning up server",
		"host", cfg.Host,
		"port", cfg.Port)
	return server.Serve(listener)
}

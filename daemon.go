package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
	"github.com/RTradeLtd/ipfs-orchestrator/delegator"

	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
)

func runDaemon(configPath string, devMode bool, args []string) {
	// load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fatal(err.Error())
	}

	println("preparing to start daemon")

	// initialize logger
	println("initializing logger")
	l, err := log.NewLogger(cfg.LogPath, devMode)
	if err != nil {
		fatal(err.Error())
	}
	defer l.Sync()
	l = l.With("version", Version)
	if cfg.LogPath != "" {
		println("logger initialized - output will be written to", cfg.LogPath)
	}

	// initialize node client
	println("initializing node client")
	c, err := ipfs.NewClient(l, cfg.IPFS)
	if err != nil {
		fatal(err.Error())
	}

	// initialize orchestrator
	println("initializing orchestrator")
	o, err := orchestrator.New(l, cfg.Address, c, cfg.IPFS.Ports, cfg.Database, devMode)
	if err != nil {
		fatal(err.Error())
	}

	// initialize daemon
	println("initializing daemon")
	dm := daemon.New(l, o)

	// initialize delegator
	println("initializing delegator")
	dl := delegator.New(l, 1*time.Minute, o.Registry)

	// catch interrupts
	ctx, cancel := context.WithCancel(context.Background())
	var signals = make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		cancel()
	}()

	// serve gRPC endpoints
	println("spinning up gRPC server...")
	go func() {
		if err := dm.Run(ctx, cfg.API); err != nil {
			println(err.Error())
		}
		cancel()
	}()

	// serve delegator
	println("spinning up delegator...")
	go func() {
		if err := dl.Run(ctx, cfg.Proxy); err != nil {
			println(err.Error())
		}
		cancel()
	}()

	// block
	<-ctx.Done()
	println("orchestrator shut down")
}

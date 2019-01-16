package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RTradeLtd/Nexus/config"
	"github.com/RTradeLtd/Nexus/daemon"
	"github.com/RTradeLtd/Nexus/ipfs"
	"github.com/RTradeLtd/Nexus/log"
	"github.com/RTradeLtd/Nexus/orchestrator"
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
	d := daemon.New(l, o)

	// handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		cancel()
	}()

	// serve endpoints
	println("spinning up server")
	if err := d.Run(ctx, cfg.API); err != nil {
		println(err.Error())
	}
	println("server shut down")
}

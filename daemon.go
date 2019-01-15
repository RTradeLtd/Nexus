package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
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

	// set up database connection
	l.Infow("intializing database connection",
		"db.host", cfg.Database.URL,
		"db.port", cfg.Database.Port,
		"db.name", cfg.Database.Name,
		"db.with_ssl", !devMode,
		"db.with_migrations", devMode)
	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: cfg.Database,
	}, database.Options{
		SSLModeDisable: devMode,
		RunMigrations:  devMode,
	})
	if err != nil {
		l.Errorw("failed to connect to database", "error", err)
		fatalf("unable to connect to database: %s", err.Error())
	}
	l.Info("successfully connected to database")
	defer func() {
		// close database
		if err := dbm.DB.Close(); err != nil {
			l.Warnw("error occurred closing database connection",
				"error", err)
		}
	}()

	// initialize orchestrator
	println("initializing orchestrator")
	o, err := orchestrator.New(l, cfg.Address, cfg.IPFS.Ports, devMode,
		c, dbm)
	if err != nil {
		fatal(err.Error())
	}

	// initialize daemon
	println("initializing daemon")
	dm := daemon.New(l, o)

	// initialize delegator
	println("initializing delegator")
	dl := delegator.New(l, Version, 1*time.Minute, []byte(cfg.Delegator.JWTKey),
		o.Registry, models.NewUserManager(dbm.DB))

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
		if err := dl.Run(ctx, cfg.Delegator); err != nil {
			println(err.Error())
		}
		cancel()
	}()

	// block
	<-ctx.Done()
	println("orchestrator shut down")
}

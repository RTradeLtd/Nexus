package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
)

var (
	configPath = flag.String("config", "./config.json", "path to ipfs-orchestrator config file")
	devMode    = flag.Bool("dev", false, "toggle dev mode")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "init":
			println("generating configuration at " + *configPath)
			config.GenerateConfig(*configPath)
			return
		case "help":
			flag.Usage()
			return
		default:
			fatal("unknown command", args[0:])
			return
		}
	}

	// load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fatal(err.Error())
	}

	// initialize logger
	l, err := log.NewLogger(*devMode)
	if err != nil {
		fatal(err.Error())
	}
	defer l.Sync()

	// initialize node client
	c, err := ipfs.NewClient(l, cfg.IPFS)
	if err != nil {
		fatal(err.Error())
	}

	// initialize orchestrator
	o, err := orchestrator.New(l, c, cfg.IPFS.Ports, cfg.Database, *devMode)
	if err != nil {
		fatal(err.Error())
	}

	// initialize daemon
	d := daemon.New(l, o)

	// handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
		os.Exit(1)
	}()

	// serve endpoints
	if err := d.Run(ctx, cfg.API); err != nil {
		println(err.Error())
	}
	println("service shut down")
	cancel()
}

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

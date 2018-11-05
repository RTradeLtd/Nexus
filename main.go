package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
)

// Version denotes the version of ipfs-orchestrator in use
var Version string

func main() {
	var (
		address    = flag.String("address", "127.0.0.1", "network address of host")
		configPath = flag.String("config", "./config.json", "path to ipfs-orchestrator config file")
		devMode    = flag.Bool("dev", os.Getenv("MODE") == "development", "toggle dev mode, alternatively MODE=development")
	)

	flag.Parse()
	if *devMode == true {
		println("[WARNING] dev mode enabled")
	}
	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "version":
			println("ipfs-orchestrator " + Version)
		case "init":
			config.GenerateConfig(*configPath)
			println("orchestrator configuration generated at " + *configPath)
			return
		case "daemon":
			runDaemon(*address, *configPath, *devMode)
			return
		default:
			fatal("unknown command", strings.Join(args[0:], " "))
			return
		}
	} else {
		fatal("no arguments provided")
	}
}

func runDaemon(address, configPath string, devMode bool) {
	// load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fatal(err.Error())
	}

	// initialize logger
	println("initializing logger")
	l, err := log.NewLogger(devMode)
	if err != nil {
		fatal(err.Error())
	}
	defer l.Sync()
	l = l.With("version", Version)

	// initialize node client
	println("initializing node client")
	c, err := ipfs.NewClient(l, cfg.IPFS)
	if err != nil {
		fatal(err.Error())
	}

	// initialize orchestrator
	println("initializing orchestrator")
	o, err := orchestrator.New(l, address, c, cfg.IPFS.Ports, cfg.Database, devMode)
	if err != nil {
		fatal(err.Error())
	}

	// initialize daemon
	println("initializing daemon")
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
	println("spinning up server")
	if err := d.Run(ctx, cfg.API); err != nil {
		println(err.Error())
	}
	println("server shut down")
	cancel()
}

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

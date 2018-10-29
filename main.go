package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
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

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fatal(err.Error())
	}

	l, err := log.NewLogger(*devMode)
	if err != nil {
		fatal(err.Error())
	}
	defer l.Sync()

	o, err := orchestrator.New(l, cfg.IPFS, cfg.Postgres)
	if err != nil {
		fatal(err.Error())
	}
	d := daemon.New(o)

	fatal(d.Run(context.Background(), cfg.API))
}

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

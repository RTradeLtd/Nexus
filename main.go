package main

import (
	"flag"
	"os"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
	"github.com/RTradeLtd/ipfs-orchestrator/orchestrator"
)

var (
	configPath = flag.String("config", "./config.json", "path to ipfs-orchestrator config file")
)

func main() {
	flag.Parse()

	if (len(os.Args) > 2) && os.Args[1] == "init" {
		println(*configPath)
		config.GenerateConfig(*configPath)
		return
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fatal(err.Error())
	}

	o, err := orchestrator.New(cfg.IPFS, cfg.Postgres)
	if err != nil {
		fatal(err.Error())
	}
	d := daemon.New(o)

	fatal(d.Run(cfg.API))
}

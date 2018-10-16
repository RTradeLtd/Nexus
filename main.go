package main

import (
	"flag"
	"os"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/daemon"
	"github.com/RTradeLtd/ipfs-orchestrator/internal"
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
		internal.Fatal(err.Error())
	}

	o, err := orchestrator.New(cfg.Postgres)
	if err != nil {
		internal.Fatal(err.Error())
	}
	d := daemon.New(o)

	internal.Fatal(d.Run(cfg.API))
}

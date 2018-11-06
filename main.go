package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
)

// Version denotes the version of ipfs-orchestrator in use
var Version string

func init() {
	if Version == "" {
		Version = "version unknown"
	}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `ipfs-orchestrator is the IPFS private network node orchestration and registry service for Temporal.

USAGE:

  ipfs-orchestrator [options] [command] [arguments...]

COMMANDS:

  init        initialize configuration
  daemon      spin up the ipfs-orchestrator daemon and processes
  ctl         [EXPERIMENTAL] interact with daemon via a low-level client
  version     display program version

OPTIONS:

`)
		flag.PrintDefaults()
	}
}

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
			runDaemon(*address, *configPath, *devMode, args[1:])
			return
		case "ctl":
			if len(args) > 1 && (args[1] == "-pretty" || args[1] == "--pretty") {
				runCTL(*configPath, *devMode, true, args[2:])
			} else {
				runCTL(*configPath, *devMode, false, args[1:])
			}
			return
		default:
			fatal(fmt.Sprintf("unknown command '%s' - run 'ipfs-orchestrator --help' for documentation", strings.Join(args[0:], " ")))
			return
		}
	} else {
		fatal("no arguments provided")
	}
}

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

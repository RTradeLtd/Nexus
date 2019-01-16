package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
)

// Version denotes the version of Nexus in use
var Version string

func init() {
	if Version == "" {
		Version = "version unknown"
	}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Nexus is the IPFS private network node orchestration and
registry service for Temporal.

USAGE:

  nexus [options] [command] [arguments...]

COMMANDS:

  init        initialize configuration
  daemon      spin up the Nexus daemon and related processes
  ctl         [EXPERIMENTAL] interact with daemon via a low-level client
  version     display program version

OPTIONS:

`)
		flag.PrintDefaults()
	}
}

func main() {
	var (
		configPath = flag.String("config", "./config.json",
			"path to Nexus configuration file")
		devMode = flag.Bool("dev", os.Getenv("MODE") == "development",
			"toggle dev mode, alternatively set using MODE=development")
	)

	flag.Parse()
	if *devMode == true {
		println("[WARNING] dev mode enabled")
	}
	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "version":
			println("Nexus " + Version)
		case "init":
			config.GenerateConfig(*configPath)
			println("orchestrator configuration generated at " + *configPath)
			return
		case "daemon":
			runDaemon(*configPath, *devMode, args[1:])
			return
		case "ctl":
			if len(args) > 1 && (args[1] == "-pretty" || args[1] == "--pretty") {
				runCTL(*configPath, *devMode, true, args[2:])
			} else {
				runCTL(*configPath, *devMode, false, args[1:])
			}
			return
		default:
			fatal(fmt.Sprintf("unknown command '%s' - user the --help' flag for documentation",
				strings.Join(args[0:], " ")))
			return
		}
	} else {
		fatal("no arguments provided - use the '--help' flag for documentation")
	}
}

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/RTradeLtd/Nexus/config"
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
	version     display program version

	dev         [DEV] utilities for development purposes
	ctl         [EXPERIMENTAL] interact with daemon via a low-level client

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
		// version info
		case "version":
			println("Nexus " + Version)
		case "init":
			config.GenerateConfig(*configPath, *devMode)
			println("orchestrator configuration generated at " + *configPath)
			return
		// run daemon
		case "daemon":
			runDaemon(*configPath, *devMode, args[1:])
			return
		// run ctl
		case "ctl":
			if len(args) > 1 && (args[1] == "-pretty" || args[1] == "--pretty") {
				runCTL(*configPath, *devMode, true, args[2:])
			} else {
				runCTL(*configPath, *devMode, false, args[1:])
			}
			return
		// dev utilities
		case "dev":
			if *devMode != true {
				fatal("do not use the dev utilities outside of dev mode!")
			}
			if len(args) > 1 {
				switch args[1] {
				case "network":
					if len(args) < 2 {
						fatal("additional argument required")
					}
					initTestNetwork(*configPath, args[2])
				default:
					fatal(fmt.Sprintf("unknown command '%s' - user the --help' flag for documentation",
						strings.Join(args[0:], " ")))
				}
			}
		// default error
		default:
			fatal(fmt.Sprintf("unknown command '%s' - user the --help' flag for documentation",
				strings.Join(args[0:], " ")))
			return
		}
	} else {
		fatal("insufficient arguments provided - use the '--help' flag for documentation")
	}
}

func fatal(msg ...interface{}) {
	fmt.Println(msg...)
	os.Exit(1)
}

func fatalf(format string, msg ...interface{}) {
	fmt.Printf(format, msg...)
	println()
	os.Exit(1)
}

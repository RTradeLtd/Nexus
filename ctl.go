package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/RTradeLtd/ctl"
	"github.com/RTradeLtd/ipfs-orchestrator/client"
	"github.com/RTradeLtd/ipfs-orchestrator/config"
)

func runCTL(configPath string, devMode, prettyPrint bool, args []string) {
	// load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fatal(err.Error())
	}

	c, err := client.New(cfg.API, devMode)
	if err != nil {
		fatal(err.Error())
	}
	defer c.Close()

	// create controller
	controller, err := ctl.New(c.ServiceClient)
	if err != nil {
		fatal(err.Error())
	}

	// show help if needed
	if args != nil && len(args) == 1 && args[0] == "help" {
		controller.Help(os.Stdout)
		return
	}

	// execute command
	start := time.Now()
	out, err := controller.Exec(context.Background(), args, os.Stdout)
	if err != nil {
		fatal(err.Error())
	}

	// parse and print output
	fmt.Printf("\nReturned after %f seconds:\n", time.Since(start).Seconds())
	if !prettyPrint {
		fmt.Printf("{ %v }\n", out)
	} else {
		b, err := json.Marshal(&out)
		if err != nil {
			fatal("failed to read output: ", err.Error())
		}
		var pretty bytes.Buffer
		if err = json.Indent(&pretty, b, "", "\t"); err != nil {
			fatal("failed to pretty-print output: ", err.Error())
		}
		println(string(append(pretty.Bytes(), '\n')))
	}
}

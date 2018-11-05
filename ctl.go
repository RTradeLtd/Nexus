package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/RTradeLtd/ctl"
	"github.com/RTradeLtd/ipfs-orchestrator/client"
	"github.com/RTradeLtd/ipfs-orchestrator/config"
)

func runCTL(configPath string, devMode bool, args []string) {
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
	out, err := controller.Exec(args, os.Stdout)
	if err != nil {
		fatal(err.Error())
	}

	// parse and print output
	fmt.Printf("%v\n", out)
	b, err := json.Marshal(&out)
	if err != nil {
		fatal(err.Error())
	}
	var pretty bytes.Buffer
	if err = json.Indent(&pretty, b, "", "\t"); err != nil {
		fatal(err.Error())
	}
	println(append(pretty.Bytes(), '\n'))
}

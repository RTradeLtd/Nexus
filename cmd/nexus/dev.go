package main

import (
	"github.com/RTradeLtd/Nexus/config"
	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
)

func initTestNetwork(configPath, networkName string) {
	// load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fatal(err.Error())
	}

	println("setting up database entry for a test network")

	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: cfg.Database,
	}, database.Options{
		SSLModeDisable: true,
		RunMigrations:  true,
	})
	if err != nil {
		fatal(err.Error())
	}

	var nm = models.NewHostedIPFSNetworkManager(dbm.DB)
	if _, err := nm.CreateHostedPrivateNetwork(networkName, "", nil, models.NetworkAccessOptions{
		Users:         []string{"testuser"},
		PublicGateway: true,
	}); err != nil {
		fatal(err.Error())
	}
}

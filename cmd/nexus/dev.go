package main

import (
	"github.com/RTradeLtd/Nexus/config"
	tcfg "github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/database/v2/models"
)

func initTestNetwork(configPath, networkName string) {
	// load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fatal(err.Error())
	}

	println("setting up database entry for a test network")

	dbm, err := database.New(&tcfg.TemporalConfig{
		Database: cfg.Database,
	}, database.Options{
		SSLModeDisable: true,
		RunMigrations:  true,
	})
	if err != nil {
		fatal(err.Error())
	}

	var nm = models.NewHostedNetworkManager(dbm.DB)
	if _, err := nm.CreateHostedPrivateNetwork(networkName, "", nil, models.NetworkAccessOptions{
		Users:         []string{"testuser"},
		PublicGateway: true,
	}); err != nil {
		fatal(err.Error())
	}
}

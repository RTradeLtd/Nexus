package orchestrator

import (
	"reflect"
	"testing"

	"github.com/RTradeLtd/Nexus/ipfs"

	tcfg "github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	"github.com/RTradeLtd/database/v2/models"
)

var dbDefaults = tcfg.Database{
	Name:     "temporal",
	URL:      "127.0.0.1",
	Port:     "5433",
	Username: "postgres",
	Password: "password123",
}

func newTestDB() (*database.Manager, error) {
	return database.New(&tcfg.TemporalConfig{
		Database: dbDefaults,
	}, database.Options{
		SSLModeDisable: true,
		RunMigrations:  true,
	})
}

func Test_getOptionsFromDatabaseEntry(t *testing.T) {
	type args struct {
		network *models.HostedNetwork
	}
	tests := []struct {
		name    string
		args    args
		want    ipfs.NodeOpts
		wantErr bool
	}{
		{"invalid network", args{nil}, ipfs.NodeOpts{}, true},
		{"with swarm key", args{&models.HostedNetwork{
			SwarmKey: "helloworld",
		}}, ipfs.NodeOpts{
			SwarmKey: []byte("helloworld"),
		}, false},
		{"without swarm key", args{&models.HostedNetwork{}}, ipfs.NodeOpts{
			SwarmKey: []byte("generated"),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getOptionsFromDatabaseEntry(tt.args.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOptionsFromDatabaseEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(tt.want.SwarmKey) == "generated" {
				if got.SwarmKey == nil || len(got.SwarmKey) == 0 {
					t.Errorf("getOptionsFromDatabaseEntry() = %v, want generated swarm key", got)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOptionsFromDatabaseEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

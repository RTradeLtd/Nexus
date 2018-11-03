package orchestrator

import (
	"reflect"
	"testing"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
)

func Test_getOptionsFromDatabaseEntry(t *testing.T) {
	type args struct {
		network *models.HostedIPFSPrivateNetwork
	}
	tests := []struct {
		name    string
		args    args
		want    ipfs.NodeOpts
		wantErr bool
	}{
		{"invalid network", args{nil}, ipfs.NodeOpts{}, true},
		{"with swarm key", args{&models.HostedIPFSPrivateNetwork{
			SwarmKey: "helloworld",
		}}, ipfs.NodeOpts{
			SwarmKey: []byte("helloworld"),
		}, false},
		{"without swarm key", args{&models.HostedIPFSPrivateNetwork{}}, ipfs.NodeOpts{
			SwarmKey: []byte("generated"),
		}, false},
		{"with bootstrap peers", args{&models.HostedIPFSPrivateNetwork{
			SwarmKey:               "helloworld",
			BootstrapPeerAddresses: []string{"1234", "5678"},
		}}, ipfs.NodeOpts{
			SwarmKey:       []byte("helloworld"),
			BootstrapPeers: []string{"1234", "5678"},
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

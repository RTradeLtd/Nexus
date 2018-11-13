package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/ipfs/mock"

	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
)

func TestNew(t *testing.T) {
	type args struct {
		pgOpts tcfg.Database
	}
	tests := []struct {
		name          string
		args          args
		wantClientErr bool
		wantErr       bool
	}{
		{"node client err", args{dbDefaults}, true, true},
		{"invalid db options", args{tcfg.Database{}}, false, true},
		{"all good", args{dbDefaults}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := log.NewTestLogger()
			client := &mock.FakeNodeClient{}

			if tt.wantClientErr {
				client.NodesReturns(nil, errors.New("oh no"))
			} else {
				client.NodesReturns([]*ipfs.NodeInfo{}, nil)
			}

			_, err := New(l, "", client, config.Ports{}, tt.args.pgOpts, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestOrchestrator_Run(t *testing.T) {
	l, _ := log.NewTestLogger()
	client := &mock.FakeNodeClient{}

	o, err := New(l, "", client, config.Ports{}, dbDefaults, true)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	o.Run(ctx)
	time.Sleep(1 * time.Millisecond)
	cancel()
}

func TestOrchestrator_NetworkUp(t *testing.T) {
	// pre-test database setup
	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: dbDefaults,
	}, database.Options{
		RunMigrations:  true,
		SSLModeDisable: true,
	})
	if err != nil {
		t.Fatalf("failed to connect to dev database: %s", err.Error())
	}
	nm := models.NewHostedIPFSNetworkManager(dbm.DB)
	testNetwork := &models.HostedIPFSPrivateNetwork{
		Name: "test-network-1",
	}
	if check := nm.DB.Create(testNetwork); check.Error != nil {
		t.Log(check.Error.Error())
	}
	defer nm.DB.Delete(testNetwork)

	type fields struct {
		regPorts config.Ports
	}
	type args struct {
		network string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		createErr bool
		wantErr   bool
	}{
		{"invalid network name", fields{config.Ports{}}, args{""}, false, true},
		{"nonexistent network", fields{config.Ports{}}, args{"asdf"}, false, true},
		{"unable to register network", fields{config.Ports{}}, args{"test-network-1"}, false, true},
		{"instantiate node with error", fields{config.New().Ports}, args{"test-network-1"}, true, true},
		{"success", fields{config.New().Ports}, args{"test-network-1"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := log.NewTestLogger()
			client := &mock.FakeNodeClient{}
			o := &Orchestrator{
				l:       l,
				nm:      nm,
				client:  client,
				reg:     registry.New(l, tt.fields.regPorts),
				address: "127.0.0.1",
			}

			if tt.createErr {
				client.CreateNodeReturns(errors.New("oh no"))
			}

			if _, err := o.NetworkUp(context.Background(), tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("Orchestrator.NetworkUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrchestrator_NetworkDown(t *testing.T) {
	// pre-test database setup
	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: dbDefaults,
	}, database.Options{
		RunMigrations:  true,
		SSLModeDisable: true,
	})
	if err != nil {
		t.Fatalf("failed to connect to dev database: %s", err.Error())
	}
	nm := models.NewHostedIPFSNetworkManager(dbm.DB)
	testNetwork := &models.HostedIPFSPrivateNetwork{
		Name: "test-network-1",
	}
	if check := nm.DB.Create(testNetwork); check.Error != nil {
		t.Log(check.Error.Error())
	}
	defer nm.DB.Delete(testNetwork)

	type fields struct {
		node ipfs.NodeInfo
	}
	type args struct {
		network string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		createErr bool
		wantErr   bool
	}{
		{"invalid network name", fields{ipfs.NodeInfo{}}, args{""}, false, true},
		{"unregistered network", fields{ipfs.NodeInfo{}}, args{"asdf"}, false, true},
		{"nonexistent network", fields{ipfs.NodeInfo{NetworkID: "asdf"}},
			args{"asdf"}, false, true},
		{"stop node with error", fields{ipfs.NodeInfo{NetworkID: "test-network-1"}},
			args{"test-network-1"}, true, false},
		{"stop node without error", fields{ipfs.NodeInfo{NetworkID: "test-network-1"}},
			args{"test-network-1"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := log.NewTestLogger()
			client := &mock.FakeNodeClient{}
			o := &Orchestrator{
				l:       l,
				nm:      nm,
				client:  client,
				reg:     registry.New(l, config.New().Ports, &tt.fields.node),
				address: "127.0.0.1",
			}

			if tt.createErr {
				client.StopNodeReturns(errors.New("oh no"))
			}

			if err := o.NetworkDown(context.Background(), tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("Orchestrator.NetworkDown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrchestrator_NetworkStatus(t *testing.T) {
	type fields struct {
		node ipfs.NodeInfo
	}
	type args struct {
		network string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		createErr bool
		wantErr   bool
	}{
		{"invalid network name", fields{ipfs.NodeInfo{}}, args{""}, false, true},
		{"unable to find node", fields{ipfs.NodeInfo{}}, args{"asdf"}, false, true},
		{"client fail", fields{ipfs.NodeInfo{NetworkID: "asdf"}}, args{"asdf"}, true, true},
		{"client succeed", fields{ipfs.NodeInfo{NetworkID: "asdf"}}, args{"asdf"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := log.NewTestLogger()
			client := &mock.FakeNodeClient{}
			o := &Orchestrator{
				l:       l,
				client:  client,
				reg:     registry.New(l, config.New().Ports, &tt.fields.node),
				address: "127.0.0.1",
			}

			if tt.createErr {
				client.NodeStatsReturns(ipfs.NodeStats{}, errors.New("oh no"))
			}

			if _, err := o.NetworkStatus(context.Background(), tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("Orchestrator.NetworkStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

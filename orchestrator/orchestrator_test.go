package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/RTradeLtd/ipfs-orchestrator/ipfs"
	ipfsmock "github.com/RTradeLtd/ipfs-orchestrator/ipfs/mock"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
	"github.com/RTradeLtd/ipfs-orchestrator/registry"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

var dbDefaults = tcfg.Database{
	Name:     "temporal",
	URL:      "127.0.0.1",
	Port:     "5433",
	Username: "postgres",
	Password: "password123",
}

func newTestIPFSClient(l *zap.SugaredLogger, t *testing.T) (*ipfsmock.MockNodeClient, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	return ipfsmock.NewMockNodeClient(ctrl), ctrl
}

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
			mock, ctrl := newTestIPFSClient(l, t)
			defer ctrl.Finish()

			if tt.wantClientErr {
				mock.EXPECT().
					Nodes(gomock.Any()).
					DoAndReturn(func(...interface{}) (interface{}, error) {
						return nil, errors.New("oh no")
					}).
					Times(1)
			} else {
				mock.EXPECT().
					Nodes(gomock.Any()).
					Times(1)
			}

			_, err := New(l, "", mock, config.Ports{}, tt.args.pgOpts, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestOrchestrator_Run(t *testing.T) {
	l, _ := log.NewTestLogger()
	mock, ctrl := newTestIPFSClient(l, t)
	defer ctrl.Finish()

	mock.EXPECT().
		Nodes(gomock.Any()).
		Times(1)

	o, err := New(l, "", mock, config.Ports{}, dbDefaults, true)
	if err != nil {
		t.Error(err)
		return
	}

	mock.EXPECT().
		Watch(gomock.Any()).
		Times(1)

	ctx, cancel := context.WithCancel(context.Background())
	o.Run(ctx)
	time.Sleep(1 * time.Millisecond)
	cancel()
}

func TestOrchestrator_NetworkUp(t *testing.T) {
	// pre-test database setup
	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: dbDefaults,
	}, database.DatabaseOptions{
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
			mock, ctrl := newTestIPFSClient(l, t)
			defer ctrl.Finish()
			o := &Orchestrator{
				l:      l,
				nm:     nm,
				client: mock,
				reg:    registry.New(l, tt.fields.regPorts),
				host:   "127.0.0.1",
			}

			if tt.createErr {
				mock.EXPECT().
					CreateNode(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("oh no")).
					Times(1)
			} else if !tt.wantErr {
				mock.EXPECT().
					CreateNode(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			}

			if err := o.NetworkUp(context.Background(), tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("Orchestrator.NetworkUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrchestrator_NetworkDown(t *testing.T) {
	// pre-test database setup
	dbm, err := database.Initialize(&tcfg.TemporalConfig{
		Database: dbDefaults,
	}, database.DatabaseOptions{
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
			mock, ctrl := newTestIPFSClient(l, t)
			defer ctrl.Finish()
			o := &Orchestrator{
				l:      l,
				nm:     nm,
				client: mock,
				reg:    registry.New(l, config.New().Ports, &tt.fields.node),
				host:   "127.0.0.1",
			}

			if tt.createErr {
				mock.EXPECT().
					StopNode(gomock.Any(), gomock.Any()).
					Return(errors.New("oh no")).
					Times(1)
			} else {
				mock.EXPECT().
					StopNode(gomock.Any(), gomock.Any()).
					AnyTimes()
			}

			if err := o.NetworkDown(context.Background(), tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("Orchestrator.NetworkDown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	tcfg "github.com/RTradeLtd/config"
	"github.com/RTradeLtd/ipfs-orchestrator/config"
	ipfsmock "github.com/RTradeLtd/ipfs-orchestrator/ipfs/mock"
	"github.com/RTradeLtd/ipfs-orchestrator/log"
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

			_, err := New(l, mock, config.Ports{}, tt.args.pgOpts, true)
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

	o, err := New(l, mock, config.Ports{}, dbDefaults, true)
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

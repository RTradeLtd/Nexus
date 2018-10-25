package log

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	type args struct {
		dev bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"dev", args{true}, false},
		{"prod", args{false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSugar, err := NewLogger(tt.args.dev)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotSugar == nil {
				t.Error("got unexpected nil logger")
			}
		})
	}
}

func TestNewTestLogger(t *testing.T) {
	logger, out := NewTestLogger()
	logger.Info("hi")
	if out.All()[0].Message != "hi" {
		t.Error("bad logger")
	}
}

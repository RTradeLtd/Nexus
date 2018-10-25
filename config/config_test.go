package config

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	def := IPFSOrchestratorConfig{}
	def.setDefaults()
	type args struct {
		configPath string
	}
	tests := []struct {
		name    string
		args    args
		want    IPFSOrchestratorConfig
		wantErr bool
	}{
		{"invalid path", args{""}, IPFSOrchestratorConfig{}, true},
		{"invalid file", args{"./config.go"}, IPFSOrchestratorConfig{}, true},
		{"valid dev config", args{"../dev/config.json"}, def, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(tt.args.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	if err := GenerateConfig("../config.json"); err != nil {
		t.Error(err.Error())
	}
}

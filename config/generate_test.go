package config

import "testing"

func TestGenerateConfig(t *testing.T) {
	if err := GenerateConfig("../config.json"); err != nil {
		t.Error(err.Error())
	}
}

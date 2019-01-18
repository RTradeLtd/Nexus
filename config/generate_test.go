package config

import "testing"

func TestGenerateConfig(t *testing.T) {
	if err := GenerateConfig("../config.json", false); err != nil {
		t.Error(err.Error())
	}
	if err := GenerateConfig("../config.json", true); err != nil {
		t.Error(err.Error())
	}
}

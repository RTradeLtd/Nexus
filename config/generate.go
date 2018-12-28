package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

// GenerateConfig writes an empty orchestrator config template to given filepath
func GenerateConfig(configPath string) error {
	template := &IPFSOrchestratorConfig{}
	template.setDefaults()
	b, err := json.Marshal(template)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err = json.Indent(&pretty, b, "", "  "); err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, append(pretty.Bytes(), '\n'), os.ModePerm)
}

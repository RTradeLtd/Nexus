package ipfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	internal "github.com/RTradeLtd/Nexus/ipfs/internal"
)

// GoIPFSConfig is a subset of go-ipfs's configuration structure
type GoIPFSConfig struct {
	Identity struct {
		PeerID  string
		PrivKey string
	}
}

func getConfig(path string) (*GoIPFSConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open IPFS configuration at '%s': %s", path, err.Error())
	}
	var c GoIPFSConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("error encountered reading contents of configuration at '%s': '%s'",
			path, err.Error())
	}
	return &c, nil
}

func newNodeStartScript(diskMax int) (string, error) {
	f, err := internal.ReadFile("ipfs/internal/ipfs_start.sh")
	if err != nil {
		return "", fmt.Errorf("failed to generate startup script: %s", err.Error())
	}
	return fmt.Sprintf(string(f),
		diskMax,
	), nil
}

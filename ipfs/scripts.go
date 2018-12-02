package ipfs

import (
	"fmt"

	internal "github.com/RTradeLtd/ipfs-orchestrator/ipfs/internal"
)

func newNodeStartScript(diskMax int) (string, error) {
	f, err := internal.ReadFile("ipfs/internal/ipfs_start.sh")
	if err != nil {
		return "", fmt.Errorf("failed to generate startup script: %s", err.Error())
	}
	return fmt.Sprintf(string(f),
		diskMax,
	), nil
}

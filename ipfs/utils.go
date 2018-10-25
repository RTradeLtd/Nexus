package ipfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirEnv = "DATA_DIR"
)

func getDataDir(network string) string {
	return filepath.Join(os.Getenv(dirEnv), fmt.Sprintf("/data/ipfs/%s", network))
}

func isNodeContainer(imageName string) bool {
	parts := strings.Split(imageName, "-")
	return len(parts) > 0 && strings.Contains(parts[0], "ipfs")
}

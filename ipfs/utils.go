package ipfs

import (
	"os"
	"path/filepath"
	"strings"
)

func toNodeContainerName(network string) string {
	return "ipfs-" + network
}

func isNodeContainer(imageName string) bool {
	parts := strings.Split(imageName, "-")
	return len(parts) > 0 && strings.Contains(parts[0], "ipfs")
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func isStopped(status string) bool {
	return status == "exited" || status == "dead"
}

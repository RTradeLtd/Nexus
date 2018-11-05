package ipfs

import (
	"strings"
)

func isNodeContainer(imageName string) bool {
	parts := strings.Split(imageName, "-")
	return len(parts) > 0 && strings.Contains(parts[0], "ipfs")
}

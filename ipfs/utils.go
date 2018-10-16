package ipfs

import (
	"fmt"
	"strings"
)

func getDataDir(network string) string { return fmt.Sprintf("/data/ipfs/%s", network) }

func parseNetworkName(imageName string) string {
	return strings.Join(strings.Split(imageName, "-")[1:], "-")
}

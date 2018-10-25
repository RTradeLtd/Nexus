package ipfs

import "testing"

func TestNode(t *testing.T) {
	// super simple tests to get coverage on node private data getters
	n := NodeInfo{}
	n.DockerID()
	n.ContainerName()
	n.DataDirectory()
	n.BootstrapPeers()
}

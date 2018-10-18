package ipfs

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	Network string
	Ports   NodePorts

	dockerID string
}

// NodePorts declares the exposed ports of an IPFS node
type NodePorts struct {
	Swarm   string // default: 4001
	API     string // default: 5001
	Gateway string // default: 8080
}

// DockerID is the ID of the node's Docker container
func (n *NodeInfo) DockerID() string { return n.dockerID }

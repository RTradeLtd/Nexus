package ipfs

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	Network string
	Ports   NodePorts

	dockerID string
}

// NodePorts declares the exposed ports of an IPFS node
type NodePorts struct {
	Swarm   string
	API     string
	Gateway string
}

// DockerID is the ID of the node's Docker container
func (n *NodeInfo) DockerID() string { return n.dockerID }

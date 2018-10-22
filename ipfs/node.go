package ipfs

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	Network string
	Ports   NodePorts

	// private metadata set by node client - access via getters
	dockerID       string
	containerName  string
	dataDir        string
	bootstrapPeers []string
}

// NodePorts declares the exposed ports of an IPFS node
type NodePorts struct {
	Swarm   string // default: 4001
	API     string // default: 5001
	Gateway string // default: 8080
}

// DockerID is the ID of the node's Docker container
func (n *NodeInfo) DockerID() string { return n.dockerID }

// ContainerName is the name of the node's Docker container
func (n *NodeInfo) ContainerName() string { return n.containerName }

// DataDirectory is the path to the directory holding all data relevant to this
// IPFS node
func (n *NodeInfo) DataDirectory() string { return n.dataDir }

// BootstrapPeers lists the peers this node was bootstrapped onto upon init
func (n *NodeInfo) BootstrapPeers() []string { return n.bootstrapPeers }

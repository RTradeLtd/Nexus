package ipfs

import (
	"encoding/json"
	"fmt"
)

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	NetworkID string
	Ports     NodePorts
	JobID     string

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

func newNode(id, name string, attributes map[string]string) (NodeInfo, error) {
	// check if container is a node
	if !isNodeContainer(name) {
		return NodeInfo{dockerID: id, containerName: name}, fmt.Errorf("unknown name format %s", name)
	}

	// parse bootstrap state
	var peers []string
	json.Unmarshal([]byte(attributes["bootstrap_peers"]), &peers)

	// create node metadata to return
	return NodeInfo{
		NetworkID: attributes["network_id"],
		Ports: NodePorts{
			Swarm:   attributes["swarm_port"],
			API:     attributes["api_port"],
			Gateway: attributes["gateway_port"],
		},
		JobID:          attributes["job_id"],
		dockerID:       id,
		containerName:  name,
		dataDir:        attributes["data_dir"],
		bootstrapPeers: peers,
	}, nil
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

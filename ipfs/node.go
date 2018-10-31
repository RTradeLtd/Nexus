package ipfs

import (
	"encoding/json"
	"fmt"
)

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	NetworkID string    `json:"network_id"`
	Ports     NodePorts `json:"ports"`
	JobID     string    `json:"job_id"`

	// Metadata set by node client:
	// DockerID is the ID of the node's Docker container
	DockerID string `json:"docker_id"`
	// ContainerName is the name of the node's Docker container
	ContainerName string `json:"container_id"`
	// DataDir is the path to the directory holding all data relevant to this
	// IPFS node
	DataDir string `json:"data_dir"`
	// BootstrapPeers lists the peers this node was bootstrapped onto upon init
	BootstrapPeers []string `json:"bootstrap_peers"`
}

// NodePorts declares the exposed ports of an IPFS node
type NodePorts struct {
	Swarm   string `json:"swarm"`   // default: 4001
	API     string `json:"api"`     // default: 5001
	Gateway string `json:"gateway"` // default: 8080
}

func newNode(id, name string, attributes map[string]string) (NodeInfo, error) {
	// check if container is a node
	if !isNodeContainer(name) {
		return NodeInfo{DockerID: id, ContainerName: name}, fmt.Errorf("unknown name format %s", name)
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
		DockerID:       id,
		ContainerName:  name,
		DataDir:        attributes["data_dir"],
		BootstrapPeers: peers,
	}, nil
}

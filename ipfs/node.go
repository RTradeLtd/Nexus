package ipfs

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	keyNetworkID = "network_id"
	keyJobID     = "job_id"

	keyBootstrapPeers = "bootstrap_peers"
	keyDataDir        = "data_dir"

	keyPortSwarm   = "ports.swarm"
	keyPortAPI     = "ports.api"
	keyPortGateway = "ports.gateway"

	keyResourcesDisk   = "resources.disk"
	keyResourcesMemory = "resources.memory"
	keyResourcesCPUs   = "resources.cpus"
)

// NodeInfo defines metadata about an IPFS node
type NodeInfo struct {
	NetworkID string `json:"network_id"`
	JobID     string `json:"job_id"`

	Ports     NodePorts     `json:"ports"`
	Resources NodeResources `json:"resources"`

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

// NodeResources declares resource quotas for this node
type NodeResources struct {
	DiskGB   int `json:"disk"`
	MemoryGB int `json:"memory"`
	CPUs     int `json:"cpus"`
}

func newNode(id, name string, attributes map[string]string) (NodeInfo, error) {
	// check if container is a node
	if !isNodeContainer(name) {
		return NodeInfo{DockerID: id, ContainerName: name}, fmt.Errorf("unknown name format %s", name)
	}

	// parse bootstrap state
	var peers []string
	json.Unmarshal([]byte(attributes[keyBootstrapPeers]), &peers)

	// parse resource data
	var (
		disk, _ = strconv.Atoi(attributes[keyResourcesDisk])
		mem, _  = strconv.Atoi(attributes[keyResourcesMemory])
		cpus, _ = strconv.Atoi(attributes[keyResourcesCPUs])
	)

	// create node metadata to return
	return NodeInfo{
		NetworkID: attributes[keyNetworkID],
		JobID:     attributes[keyJobID],

		Ports: NodePorts{
			Swarm:   attributes[keyPortSwarm],
			API:     attributes[keyPortAPI],
			Gateway: attributes[keyPortGateway],
		},
		Resources: NodeResources{
			DiskGB:   disk,
			MemoryGB: mem,
			CPUs:     cpus,
		},

		DockerID:       id,
		ContainerName:  name,
		DataDir:        attributes[keyDataDir],
		BootstrapPeers: peers,
	}, nil
}

func (n *NodeInfo) withDefaults() {
	if n.Resources.CPUs == 0 {
		n.Resources.CPUs = 4
	}
	if n.Resources.DiskGB == 0 {
		n.Resources.DiskGB = 100
	}
	if n.Resources.MemoryGB == 0 {
		n.Resources.MemoryGB = 4
	}
}

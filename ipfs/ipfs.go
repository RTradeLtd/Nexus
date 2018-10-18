package ipfs

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// NodeClient provides an interface to the base Docker client for controlling
// IPFS nodes
type NodeClient interface {
	// Nodes retrieves a list of active IPFS ndoes
	Nodes(ctx context.Context) (nodes []*NodeInfo, err error)

	// CreateNode activates a new IPFS node
	CreateNode(ctx context.Context, n *NodeInfo, opts NodeOpts) (err error)

	// StopNode shuts down an existing IPFS node
	StopNode(ctx context.Context, n *NodeInfo) (err error)
}

type client struct {
	ipfsImage string

	d *docker.Client
}

// NewClient creates a new Docker Client from ENV values and negotiates the
// correct API version to use
func NewClient(ipfsOpts config.IPFS) (NodeClient, error) {
	ipfsImage := "ipfs/go-ipfs:" + ipfsOpts.Version
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	_, err = d.ImagePull(context.Background(), ipfsImage, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download IPFS image: %s", err.Error())
	}

	return &client{ipfsImage, d}, nil
}

func (c *client) Nodes(ctx context.Context) ([]*NodeInfo, error) {
	ctrs, err := c.d.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]*NodeInfo, 0)
	for _, container := range ctrs {
		// parse ports
		nodePorts := NodePorts{}
		for _, cp := range container.Ports {
			switch cp.PrivatePort {
			case 4001:
				nodePorts.Swarm = string(cp.PublicPort)
			case 5001:
				nodePorts.API = string(cp.PublicPort)
			case 8080:
				nodePorts.Gateway = string(cp.PublicPort)
			}
		}

		// create node metadata to return
		nodes = append(nodes, &NodeInfo{
			Network:  parseNetworkName(container.Names[0]),
			Ports:    nodePorts,
			dockerID: container.ID,
		})
	}

	return nodes, nil
}

// NodeOpts declares options for starting up nodes
type NodeOpts struct {
	SwarmKey       []byte
	BootstrapNodes []string
}

func (c *client) CreateNode(ctx context.Context, n *NodeInfo, opts NodeOpts) error {
	if n == nil || n.Network == "" || opts.SwarmKey == nil {
		return errors.New("invalid configuration provided")
	}

	// write swarm.key to mount point
	if err := ioutil.WriteFile(
		getConfigDir(n.Network)+"/swarm.key",
		opts.SwarmKey, 0755,
	); err != nil {
		return err
	}

	var (
		ports = nat.PortMap{
			// TODO: do these all ports need to be public?
			"4001": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: n.Ports.Swarm}},
			"5001": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: n.Ports.API}},
			"8080": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: n.Ports.Gateway}},
		}
		volumes = []string{
			getDataDir(n.Network) + ":/data/ipfs",
			getConfigDir(n.Network) + ":/config/.ipfs",
		}
		labels = map[string]string{
			"network_name": n.Network,
			"data_dir":     getDataDir(n.Network),
			"swarm_port":   n.Ports.Swarm,
			"api_port":     n.Ports.API,
			"gateway_port": n.Ports.Gateway,
			"bootstrapped": strconv.FormatBool(opts.BootstrapNodes != nil && len(opts.BootstrapNodes) > 0),
		}
	)

	// create container
	resp, err := c.d.ContainerCreate(
		ctx,
		&container.Config{
			Image: c.ipfsImage,
			Cmd: []string{
				"daemon", "--migrate=true", "--enable-pubsub-experiment",
			},
			Env: []string{
				"LIBP2P_FORCE_PNET=1", // enforce private networks
				"IPFS_PATH=/config/.ipfs",
			},
			Labels: labels,
		},
		&container.HostConfig{
			Binds:        volumes,
			PortBindings: ports,

			// TODO: limit resources
			Resources: container.Resources{},
		},
		nil, "ipfs-"+n.Network,
	)
	if err != nil {
		return fmt.Errorf("failed to instantiate node: %s", err.Error())
	}
	n.dockerID = resp.ID

	// check for warnings
	if len(resp.Warnings) > 0 {
		return errors.New(strings.Join(resp.Warnings, "\n"))
	}

	// spin up node
	if err := c.d.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *client) StopNode(ctx context.Context, n *NodeInfo) error {
	if n == nil {
		return errors.New("invalid node")
	}

	// stop container
	timeout := time.Duration(10 * time.Second)
	if err := c.d.ContainerStop(ctx, n.DockerID(), &timeout); err != nil {
		return err
	}

	// remove ipfs data
	return os.RemoveAll(getDataDir(n.Network))
}

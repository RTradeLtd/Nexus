package ipfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"

	"github.com/RTradeLtd/Nexus/log"
	"github.com/RTradeLtd/Nexus/network"
)

// Client is the primary implementation of the NodeClient interface. Instantiate
// using ipfs.NewClient()
type Client struct {
	l *zap.SugaredLogger
	d *docker.Client

	ipfsImage string
	dataDir   string
	fileMode  os.FileMode
}

// Nodes retrieves a list of active IPFS ndoes
func (c *Client) Nodes(ctx context.Context) ([]*NodeInfo, error) {
	ctrs, err := c.d.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	// small function to restart a stopped node
	restartNode := func(node NodeInfo) error {
		var l = c.l.With("node", node)
		l.Infow("restarting stopped node")
		if err := c.CreateNode(ctx, &node, NodeOpts{}); err != nil {
			l.Errorw("failed to restart node",
				"error", err)
			return err
		}
		return nil
	}

	// parse node data, restarting stopped containers if necessary
	var (
		nodes    = make([]*NodeInfo, 0)
		ignored  = 0
		restarts = 0
		failed   = 0
	)
	for _, container := range ctrs {
		var l = c.l.With("container.id", container.ID, "container.name", container.Names[0])
		n, err := newNode(container.ID, container.Names[0], container.Labels)
		if err != nil {
			l.Debugw("container ignored", "reason", err)
			ignored++
			continue
		}
		l = l.With("node", n)
		if isStopped(container.State) {
			if err := restartNode(n); err != nil {
				l.Errorw("node container failed to restart - removing", "error", err)
				if err := c.StopNode(ctx, &n); err != nil {
					l.Warn("failed to stop node", "error", err)
				}
				failed++
				continue
			}
			restarts++
		}
		nodes = append(nodes, &n)
	}

	// report activity
	c.l.Infow("all nodes checked",
		"found", len(ctrs),
		"valid", len(nodes),
		"ignored", ignored,
		"restarts", restarts,
		"failed_restarts", failed)

	return nodes, nil
}

// NodeOpts declares options for starting up nodes
type NodeOpts struct {
	SwarmKey   []byte
	AutoRemove bool
}

// CreateNode activates a new IPFS node
func (c *Client) CreateNode(ctx context.Context, n *NodeInfo, opts NodeOpts) error {
	if n == nil || n.NetworkID == "" {
		return errors.New("invalid configuration provided")
	}

	// make sure important fields are all populated
	n.withDefaults()

	// set up logger to record process events
	var l = log.NewProcessLogger(c.l, "create_node",
		"network_id", n.NetworkID)

	// initialize node assets, such as swarm keys and startup scripts
	if err := c.initNodeAssets(n, opts); err != nil {
		l.Warnw("failed to init filesystem for node", "error", err)
		return fmt.Errorf("failed to set up filesystem for node: %s", err.Error())
	}

	// set up basic configuration
	var (
		ports = nat.PortMap{
			// TODO: make this private - blocked by lack of multiaddr support for /http
			// paths, which means delegator can't work with go-ipfs swarm.
			// See https://github.com/multiformats/multiaddr/issues/63
			"4001/tcp": []nat.PortBinding{{HostIP: network.Public, HostPort: n.Ports.Swarm}},

			// API server connections can be made via delegator. Suffers from same
			// issue as above, but direct API exposure is dangeorous since it is
			// authenticated. Delegator can handle authentication
			"5001/tcp": []nat.PortBinding{{HostIP: network.Private, HostPort: n.Ports.API}},

			// Gateway connections can be made via delegator, with access controlled
			// by database
			"8080/tcp": []nat.PortBinding{{HostIP: network.Private, HostPort: n.Ports.Gateway}},
		}
		volumes = []string{
			c.getDataDir(n.NetworkID) + ":/data/ipfs",
			c.getDataDir(n.NetworkID) + "/ipfs_start:/usr/local/bin/start_ipfs",
		}
		restartPolicy = container.RestartPolicy{Name: "unless-stopped"}

		// important metadata about node
		labels = n.labels(n.BootstrapPeers, c.getDataDir(n.NetworkID))
	)

	// remove restart policy if AutoRemove is enabled
	if opts.AutoRemove {
		restartPolicy = container.RestartPolicy{}
	}

	// create ipfs node container
	containerConfig := &container.Config{
		Image: c.ipfsImage,
		Cmd: []string{
			"daemon", "--migrate=true", "--enable-pubsub-experiment",
		},
		Env: []string{
			"LIBP2P_FORCE_PNET=1", // enforce private networks
		},
		Labels:       labels,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
	}
	containerHostConfig := &container.HostConfig{
		AutoRemove:    opts.AutoRemove,
		RestartPolicy: restartPolicy,
		Binds:         volumes,
		PortBindings:  ports,
		Resources:     containerResources(n),
	}

	var start = time.Now()
	l = l.With("container.name", n.ContainerName)
	l.Debugw("creating network container",
		"container.config", containerConfig,
		"container.host_config", containerHostConfig)
	resp, err := c.d.ContainerCreate(ctx, containerConfig, containerHostConfig, nil, n.ContainerName)
	if err != nil {
		l.Errorw("failed to create container",
			"error", err, "build.duration", time.Since(start))
		return fmt.Errorf("failed to instantiate node: %s", err.Error())
	}
	l = l.With("container.id", resp.ID)
	l.Infow("container created",
		"build.duration", time.Since(start))

	// check for warnings
	if len(resp.Warnings) > 0 {
		l.Warnw("warnings encountered on container build",
			"warnings", resp.Warnings)
	}

	// assign node metadata
	n.DockerID = resp.ID
	n.DataDir = c.getDataDir(n.NetworkID)

	// spin up node
	l.Info("starting container")
	start = time.Now()
	if err := c.d.ContainerStart(ctx, n.DockerID, types.ContainerStartOptions{}); err != nil {
		l.Errorw("error occurred on startup - removing container",
			"error", err, "start.duration", time.Since(start))
		go c.d.ContainerRemove(ctx, n.ContainerName, types.ContainerRemoveOptions{Force: true})
		return fmt.Errorf("failed to start ipfs node: %s", err.Error())
	}

	// wait for node to start
	if err := c.waitForNode(ctx, n.DockerID); err != nil {
		l.Errorw("error occurred waiting for IPFS daemon startup",
			"error", err, "start.duration", time.Since(start))
		return err
	}

	// bootstrap peers if required
	if len(n.BootstrapPeers) > 0 {
		l.Debugw("bootstrapping network node with provided peers")
		if err := c.bootstrapNode(ctx, n.DockerID, n.BootstrapPeers...); err != nil {
			l.Warnw("failed to bootstrap node - stopping container",
				"error", err, "start.duration", time.Since(start))
			go c.StopNode(ctx, n)
			return fmt.Errorf("failed to bootstrap network node with provided peers: %s", err.Error())
		}
	}

	// everything is good to go
	l.Infow("network container started without issue",
		"start.duration", time.Since(start))
	return nil
}

// UpdateNode updates node configuration
func (c *Client) UpdateNode(ctx context.Context, n *NodeInfo) error {
	if n.NetworkID == "" && n.DockerID == "" {
		return errors.New("network name or docker ID required")
	}

	// set defaults
	n.withDefaults()

	var (
		l     = log.NewProcessLogger(c.l, "node_update", "node", n)
		start = time.Now()

		resp container.ContainerUpdateOKBody
		err  error
	)

	// update Docker-managed configuration
	var res = containerResources(n)
	l.Debugw("updating docker-based configuration",
		"container.resources", containerResources(n))
	resp, err = c.d.ContainerUpdate(ctx, n.DockerID, container.UpdateConfig{Resources: res})
	if err != nil {
		l.Errorw("failed to update container configuration",
			"error", err, "warnings", resp.Warnings)
		return fmt.Errorf("failed to update node configuration: %s", err.Error())
	}
	if len(resp.Warnings) > 0 {
		l.Warnw("warnings encountered updating container",
			"warnings", resp.Warnings)
	}

	// update IPFS configuration - currently requires restart, see function docs
	l.Debugw("updating IPFS node configuration",
		"node.disk", n.Resources.DiskGB)
	if err = c.updateIPFSConfig(ctx, n); err != nil {
		l.Errorw("failed to update IPFS daemon configuration",
			"error", err)
		return fmt.Errorf("failed to update IPFS configuration: %s", err.Error())
	}

	l.Infow("successfully updated network node",
		"duration", time.Since(start))
	return nil
}

// StopNode shuts down an existing IPFS node
func (c *Client) StopNode(ctx context.Context, n *NodeInfo) error {
	if n == nil || n.DockerID == "" {
		return errors.New("invalid node")
	}

	var (
		start   = time.Now()
		timeout = time.Duration(10 * time.Second)

		l = c.l.With(
			"network_id", n.NetworkID,
			"docker_id", n.DockerID)
	)

	// stop container
	err1 := c.d.ContainerStop(ctx, n.DockerID, &timeout)
	if err1 != nil {
		l.Warnw("error stopping container", "error", err1)
	}

	// remove container
	err2 := c.d.ContainerRemove(ctx, n.DockerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	if err2 != nil {
		l.Warnw("error removing container", "error", err2)
	}

	// log duration
	l.Infow("node stopped",
		"shutdown.duration", time.Since(start))

	// check and return errors
	if err1 == nil || err2 == nil {
		return nil
	}
	return fmt.Errorf(
		"errors encountered: { ContainerStop: '%v', ContainerRemove: '%v' }",
		err1, err2,
	)
}

// RemoveNode removes assets for given node
func (c *Client) RemoveNode(ctx context.Context, network string) error {
	var (
		start = time.Now()
		dir   = c.getDataDir(network)
		l     = c.l.With("network_id", network, "data_dir", dir)
	)

	l.Debug("removing node assets")
	if err := os.RemoveAll(dir); err != nil {
		l.Warnw("error encountered removing node directories",
			"error", err,
			"duration", time.Since(start))
		return fmt.Errorf("error occurred while removing assets for '%s'", network)
	}

	l.Infow("node data removed",
		"duration", time.Since(start))
	return nil
}

// NodeStats provides details about a node container
type NodeStats struct {
	Uptime    time.Duration
	DiskUsage int64
	Stats     interface{}
}

// NodeStats retrieves statistics about the provided node
func (c *Client) NodeStats(ctx context.Context, n *NodeInfo) (NodeStats, error) {
	var start = time.Now()

	// retrieve details from stats API
	s, err := c.d.ContainerStats(ctx, n.DockerID, false)
	if err != nil {
		return NodeStats{}, err
	}
	defer s.Body.Close()
	b, err := ioutil.ReadAll(s.Body)
	if err != nil {
		return NodeStats{}, err
	}
	var stats rawContainerStats
	if err = json.Unmarshal(b, &stats); err != nil {
		return NodeStats{}, err
	}

	// retrieve details from container inspection
	info, err := c.d.ContainerInspect(ctx, n.DockerID)
	if err != nil {
		return NodeStats{}, err
	}
	created, err := time.Parse(time.RFC3339, info.Created)
	if err != nil {
		return NodeStats{}, err
	}

	// check disk usage
	usage, err := dirSize(n.DataDir)
	if err != nil {
		return NodeStats{}, fmt.Errorf("failed to calculate disk usage: %s", err.Error())
	}

	c.l.Debugw("retrieved node container data",
		"network_id", n.NetworkID,
		"docker_id", n.DockerID,
		"stat.duration", time.Since(start))

	return NodeStats{
		Uptime:    time.Since(created),
		Stats:     stats,
		DiskUsage: usage,
	}, nil
}

// Event is a node-related container event
type Event struct {
	Time   int64    `json:"time"`
	Status string   `json:"status"`
	Node   NodeInfo `json:"node"`
}

// Watch initializes a goroutine that tracks IPFS node events
func (c *Client) Watch(ctx context.Context) (<-chan Event, <-chan error) {
	var (
		events = make(chan Event)
		errs   = make(chan error)
	)

	go func() {
		defer close(errs)
		eventsCh, eventsErrCh := c.d.Events(ctx,
			types.EventsOptions{Filters: filters.NewArgs(
				filters.KeyValuePair{Key: "event", Value: "die"},
				filters.KeyValuePair{Key: "event", Value: "start"},
			)})

		for {
			select {
			case <-ctx.Done():
				break

			// pipe errors back
			case err := <-eventsErrCh:
				if err != nil {
					errs <- err
				}

			// report events
			case status := <-eventsCh:
				id := status.ID[:11]
				name := status.Actor.Attributes["name"]
				node, err := newNode(id, name, status.Actor.Attributes)
				if err != nil {
					c.l.Warnw("failed to parse node", "error", err)
					continue
				}
				e := Event{Time: status.Time, Status: status.Status, Node: node}
				c.l.Infow("event received",
					"event", e)
				events <- e
			}
		}
	}()

	return events, errs
}

package ipfs

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
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

	// NodeStats retrieves statistics about the provided node
	NodeStats(ctx context.Context, n *NodeInfo) (stats NodeStats, err error)

	// Watch initializes a goroutine that tracks IPFS node events
	Watch(ctx context.Context) (<-chan Event, <-chan error)
}

type client struct {
	l *zap.SugaredLogger
	d *docker.Client

	ipfsImage string
	dataDir   string
	fileMode  os.FileMode
}

// NewClient creates a new Docker Client from ENV values and negotiates the
// correct API version to use
func NewClient(logger *zap.SugaredLogger, ipfsOpts config.IPFS) (NodeClient, error) {
	d, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dockerd: %s", err.Error())
	}
	d.NegotiateAPIVersion(context.Background())

	// parse file mode - 0 allows the stdlib to decide how to parse
	mode, err := strconv.ParseUint(ipfsOpts.ModePerm, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse perm_mode %s: %s", ipfsOpts.ModePerm, err.Error())
	}

	// pull required images
	ipfsImage := "ipfs/go-ipfs:" + ipfsOpts.Version
	if _, err = d.ImagePull(context.Background(), ipfsImage, types.ImagePullOptions{}); err != nil {
		return nil, fmt.Errorf("failed to download IPFS image: %s", err.Error())
	}

	c := &client{
		l:         logger.Named("ipfs"),
		d:         d,
		ipfsImage: ipfsImage,
		dataDir:   ipfsOpts.DataDirectory,
		fileMode:  os.FileMode(mode),
	}

	// initialize directories
	os.MkdirAll(c.getDataDir(""), 0755)

	return c, nil
}

func (c *client) Nodes(ctx context.Context) ([]*NodeInfo, error) {
	ctrs, err := c.d.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	// small function to restart a stopped node
	restartNode := func(node NodeInfo) error {
		logger := c.l.With("node", node)
		logger.Infow("restarting stopped node")
		if err := c.CreateNode(ctx, &node, NodeOpts{
			BootstrapPeers: node.BootstrapPeers,
		}); err != nil {
			logger.Errorw("failed to restart node",
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
		n, err := newNode(container.ID, container.Names[0], container.Labels)
		if err != nil {
			ignored++
			continue
		}
		if isStopped(container.Status) {
			if err := restartNode(n); err != nil {
				failed++
				continue
			}
			restarts++
		}
		nodes = append(nodes, &n)
	}

	// report activity
	c.l.Infow("all nodes checked",
		"ignored", ignored,
		"found", len(nodes),
		"restarts", restarts,
		"failed_restarts", failed)

	return nodes, nil
}

// NodeOpts declares options for starting up nodes
type NodeOpts struct {
	SwarmKey       []byte
	BootstrapPeers []string
	AutoRemove     bool
}

func (c *client) CreateNode(ctx context.Context, n *NodeInfo, opts NodeOpts) error {
	if n == nil || n.NetworkID == "" || opts.SwarmKey == nil {
		return errors.New("invalid configuration provided")
	}
	logger := c.l.With("network_id", n.NetworkID)

	// set up directories
	os.MkdirAll(c.getDataDir(n.NetworkID), c.fileMode)

	// write swarm.key to mount point
	if err := ioutil.WriteFile(
		c.getDataDir(n.NetworkID)+"/swarm.key",
		opts.SwarmKey, c.fileMode,
	); err != nil {
		return fmt.Errorf("failed to write key: %s", err.Error())
	}

	// check peers
	bootstrap := opts.BootstrapPeers != nil && len(opts.BootstrapPeers) > 0
	peerBytes, _ := json.Marshal(opts.BootstrapPeers)

	// set up basic configuration
	var (
		containerName = "ipfs-" + n.NetworkID
		ports         = nat.PortMap{
			// public ports
			"4001/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: n.Ports.Swarm}},
			"5001/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: n.Ports.API}},

			// private ports
			"8080/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: n.Ports.Gateway}},
		}
		volumes = []string{
			c.getDataDir(n.NetworkID) + ":/data/ipfs",
		}
		labels = map[string]string{
			"network_id":      n.NetworkID,
			"data_dir":        c.getDataDir(n.NetworkID),
			"swarm_port":      n.Ports.Swarm,
			"api_port":        n.Ports.API,
			"gateway_port":    n.Ports.Gateway,
			"bootstrap_peers": string(peerBytes),
			"job_id":          n.JobID,
		}
		restartPolicy container.RestartPolicy
	)

	// set restart policy
	if !opts.AutoRemove {
		restartPolicy = container.RestartPolicy{
			Name: "unless-stopped",
		}
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

		// TODO: limit resources
		Resources: container.Resources{},
	}
	start := time.Now()
	logger = logger.With("container.name", containerName)
	logger.Infow("creating network container",
		"container.config", containerConfig,
		"container.host_config", containerHostConfig)
	resp, err := c.d.ContainerCreate(ctx, containerConfig, containerHostConfig, nil, containerName)
	if err != nil {
		logger.Errorw("failed to create container",
			"error", err)
		return fmt.Errorf("failed to instantiate node: %s", err.Error())
	}
	logger = logger.With("container.id", resp.ID)
	logger.Infow("container created",
		"build.duration", time.Since(start))

	// check for warnings
	if len(resp.Warnings) > 0 {
		logger.Warnw("warnings encountered on container build",
			"warnings", resp.Warnings)
	}

	// assign node metadata
	n.DockerID = resp.ID
	n.ContainerName = containerName
	n.DataDir = c.getDataDir(n.NetworkID)

	// spin up node
	logger.Info("starting container")
	start = time.Now()
	if err := c.d.ContainerStart(ctx, n.DockerID, types.ContainerStartOptions{}); err != nil {
		logger.Errorw("error occurred on startup - removing container",
			"error", err)
		go c.d.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
		return fmt.Errorf("failed to start ipfs node: %s", err.Error())
	}

	// wait for node to start
	if err := c.waitForNode(ctx, n.DockerID); err != nil {
		return err
	}

	// bootstrap peers if required
	if bootstrap {
		logger.Info("bootstrapping network node with provided peers")
		if err := c.bootstrapNode(ctx, n.DockerID, opts.BootstrapPeers...); err != nil {
			logger.Warnw("failed to bootstrap node - stopping container",
				"error", err)
			go c.StopNode(ctx, n)
			return fmt.Errorf("failed to bootstrap network node with provided peers: %s", err.Error())
		}
	}

	// everything is good to go
	logger.Infow("container started",
		"startup.duration", time.Since(start))
	return nil
}

func (c *client) StopNode(ctx context.Context, n *NodeInfo) error {
	if n == nil || n.DockerID == "" {
		return errors.New("invalid node")
	}

	logger := c.l.With(
		"network_id", n.NetworkID,
		"docker_id", n.DockerID)

	// stop container
	start := time.Now()
	timeout := time.Duration(10 * time.Second)
	err1 := c.d.ContainerStop(ctx, n.DockerID, &timeout)
	if err1 != nil {
		logger.Warnw("error stopping container", "error", err1)
	}

	// remove container
	logger.Info("removing container")
	err2 := c.d.ContainerRemove(ctx, n.DockerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	if err2 != nil {
		logger.Warnw("error removing container", "error", err2)
	}

	// remove data
	err3 := os.RemoveAll(c.getDataDir(n.NetworkID))
	if err3 != nil {
		logger.Warnw("error removing node data", "error", err3)
	}

	// log duration
	logger.Infow("node stopped",
		"shutdown.duration", time.Since(start))

	// check and return errors
	if err1 == nil || err2 == nil || err3 == nil {
		return nil
	}
	return fmt.Errorf(
		"errors encountered: { ContainerStop: '%v', ContainerRemove: '%v', os.RemoveAll: '%v'}",
		err1, err2, err3,
	)
}

// NodeStats provides details about a node container
type NodeStats struct {
	Uptime    time.Duration
	DiskUsage int64
	Stats     interface{}
}

func (c *client) NodeStats(ctx context.Context, n *NodeInfo) (NodeStats, error) {
	start := time.Now()

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

	c.l.Infow("retrieved node container data",
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

// Watch listens for specific container events
func (c *client) Watch(ctx context.Context) (<-chan Event, <-chan error) {
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

func (c *client) getDataDir(network string) string {
	p, _ := filepath.Abs(filepath.Join(c.dataDir, fmt.Sprintf("/data/ipfs/%s", network)))
	return p
}

func (c *client) waitForNode(ctx context.Context, dockerID string) error {
	logs, err := c.d.ContainerLogs(ctx, dockerID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer logs.Close()

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled wait for %s", dockerID)
		default:
			if strings.Contains(scanner.Text(), "Daemon is ready") {
				return nil
			}
		}
	}

	return scanner.Err()
}

func (c *client) bootstrapNode(ctx context.Context, dockerID string, peers ...string) error {
	if peers == nil || len(peers) == 0 {
		return errors.New("no peers provided")
	}

	// remove default peers
	rmBootstrap := []string{"ipfs", "bootstrap", "rm", "--all"}
	exec, err := c.d.ContainerExecCreate(ctx, dockerID, types.ExecConfig{Cmd: rmBootstrap})
	if err != nil {
		return err
	}
	if err := c.d.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{}); err != nil {
		return err
	}

	// bootstrap custom peers
	bootstrap := []string{"ipfs", "bootstrap", "add"}
	exec, err = c.d.ContainerExecCreate(ctx, dockerID, types.ExecConfig{
		Cmd: append(bootstrap, peers...),
	})
	if err != nil {
		return fmt.Errorf("failed to init bootstrapping process with %s: %s", dockerID, err.Error())
	}

	return c.d.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
}

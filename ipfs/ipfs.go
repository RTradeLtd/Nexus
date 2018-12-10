package ipfs

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/RTradeLtd/ipfs-orchestrator/config"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"go.uber.org/zap"
)

// NodeClient provides an interface to the base Docker client for controlling
// IPFS nodes
type NodeClient interface {
	// Nodes retrieves a list of active IPFS ndoes
	Nodes(ctx context.Context) (nodes []*NodeInfo, err error)

	// CreateNode activates a new IPFS node
	CreateNode(ctx context.Context, n *NodeInfo, opts NodeOpts) (err error)

	// UpdateNode updates node configuration
	UpdateNode(ctx context.Context, n *NodeInfo) (err error)

	// StopNode shuts down an existing IPFS node
	StopNode(ctx context.Context, n *NodeInfo) (err error)

	// RemoveNode removes assets for given node
	RemoveNode(ctx context.Context, network string) (err error)

	// NodeStats retrieves statistics about the provided node
	NodeStats(ctx context.Context, n *NodeInfo) (stats NodeStats, err error)

	// Watch initializes a goroutine that tracks IPFS node events
	Watch(ctx context.Context) (<-chan Event, <-chan error)
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

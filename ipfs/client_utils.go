package ipfs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

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

func (c *client) initNodeAssets(n *NodeInfo, opts NodeOpts) error {
	// set up directories
	os.MkdirAll(c.getDataDir(n.NetworkID), c.fileMode)

	// write swarm.key to mount point, otherwise check if a swarm key exists
	keyPath := c.getDataDir(n.NetworkID) + "/swarm.key"
	if opts.SwarmKey != nil {
		c.l.Info("writing provided swarm key to disk",
			"node.key_path", keyPath)
		if err := ioutil.WriteFile(keyPath, opts.SwarmKey, c.fileMode); err != nil {
			return fmt.Errorf("failed to write key: %s", err.Error())
		}
	} else {
		c.l.Info("no swarm key provided - attempting to find existing key",
			"node.key_path", keyPath)
		if _, err := os.Stat(keyPath); err != nil {
			return fmt.Errorf("unable to find swarm key: %s", err.Error())
		}
	}

	// generate initialization script
	script, err := newNodeStartScript(n.Resources.DiskGB)
	if err != nil {
		return fmt.Errorf("failed to generate startup script: %s", err.Error())
	}
	if err := ioutil.WriteFile(
		c.getDataDir(n.NetworkID)+"/ipfs_start",
		[]byte(script),
		c.fileMode,
	); err != nil {
		return fmt.Errorf("failed to generate startup script: %s", err.Error())
	}

	return nil
}

func (c *client) bootstrapNode(ctx context.Context, dockerID string, peers ...string) error {
	if peers == nil || len(peers) == 0 {
		return errors.New("no peers provided")
	}

	// remove default peers
	c.containerExec(ctx, dockerID, []string{"ipfs", "bootstrap", "rm", "--all"})

	// bootstrap custom peers
	return c.containerExec(ctx, dockerID,
		append([]string{"ipfs", "bootstrap", "add"}, peers...))
}

func (c *client) updateIPFSConfig(ctx context.Context, n *NodeInfo) error {
	// NOTE: ideally, we would use ipfs commands to update the daemon configuration.
	// this is currently impossible - see:
	// https://github.com/ipfs/go-ipfs/issues/4380
	/*
		if len(config) == 0 {
			return errors.New("no config provided")
		}
		for key, val := range config {
			if err := c.containerExec(ctx, dockerID, []string{"ipfs", "config", key, val}); err != nil {
				return err
			}
		}
	*/

	script, err := newNodeStartScript(n.Resources.DiskGB)
	if err != nil {
		return fmt.Errorf("failed to generate startup script: %s", err.Error())
	}
	if err := ioutil.WriteFile(
		c.getDataDir(n.NetworkID)+"/ipfs_start",
		[]byte(script),
		c.fileMode,
	); err != nil {
		return fmt.Errorf("failed to generate startup script: %s", err.Error())
	}

	var wait = 1 * time.Second
	if err := c.d.ContainerRestart(ctx, n.DockerID, &wait); err != nil {
		return fmt.Errorf("failed to restart container: %s", err.Error())
	}

	return c.waitForNode(ctx, n.DockerID)
}

func (c *client) containerExec(ctx context.Context, dockerID string, args []string) error {
	exec, err := c.d.ContainerExecCreate(ctx, dockerID, types.ExecConfig{Cmd: args})
	if err != nil {
		return err
	}
	return c.d.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
}

package ipfsapi

import (
	"fmt"
	"net/http"
	"path"
)

// Client implements a bare-bone API for IPFS nodes. Intended for accessing
// local nodes.
type Client struct {
	apiPort string
}

// New instantiates a new IPFS API client and attempts to connect
func New(apiPort string) (*Client, error) {
	var c = &Client{
		apiPort: apiPort,
	}
	if err := c.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to IPFS API: %s", err.Error())
	}
	return c, nil
}

// Ping checks the version endpoint for the purpose of testing connectivity
func (c *Client) Ping() error {
	resp, err := http.Get(c.apiPath("version"))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("received status '%d'", resp.StatusCode)
	}
	return nil
}

func (c *Client) apiPath(endpoint string) string {
	var base = fmt.Sprintf("127.0.0.1:%s/api/v0/", c.apiPort)
	return path.Join(base, endpoint)
}

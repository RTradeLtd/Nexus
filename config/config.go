package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	tcfg "github.com/RTradeLtd/config"
)

// DefaultIPFSVersion declares the current default version of go-ipfs to use
const DefaultIPFSVersion = "v0.4.18"

// IPFSOrchestratorConfig configures the orchestration daemon
type IPFSOrchestratorConfig struct {
	// Address is the address through which external clients connect to this host
	Address string `json:"address"`

	// LogPath, if given, will be where logs are written
	LogPath string `json:"log_path"`

	IPFS          `json:"ipfs"`
	API           `json:"api"`
	Delegator     `json:"delegator"`
	tcfg.Database `json:"postgres"`
}

// IPFS configures settings relevant to IPFS nodes
type IPFS struct {
	Version       string `json:"version"`
	DataDirectory string `json:"data_dir"`
	ModePerm      string `json:"perm_mode"`
	Ports         `json:"ports"`
}

// Ports declares port-range configuration for IPFS nodes. Elements of each
// array can be of the form "<PORT>" or "<LOWER>-<UPPER>"
type Ports struct {
	Swarm   []string `json:"swarm"`
	API     []string `json:"api"`
	Gateway []string `json:"gateway"`
}

// API declares configuration for the orchestrator daemon's gRPC API
type API struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
	TLS  `json:"tls"`
}

// Delegator declares configuration for the orchestrator proxy
type Delegator struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	JWTKey string `json:"jwt_key"`
	TLS    `json:"tls"`
}

// TLS declares HTTPS configuration
type TLS struct {
	CertPath string `json:"cert"`
	KeyPath  string `json:"key"`
}

// New creates a new, default configuration
func New() IPFSOrchestratorConfig {
	var cfg IPFSOrchestratorConfig
	cfg.SetDefaults(false)
	return cfg
}

// LoadConfig loads a TemporalConfig from given filepath
func LoadConfig(configPath string) (IPFSOrchestratorConfig, error) {
	var cfg IPFSOrchestratorConfig

	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("could not open config: %s", err.Error())
	}

	if err = json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("could not read config: %s", err.Error())
	}

	cfg.SetDefaults(false)

	return cfg, nil
}

// SetDefaults initializes certain blank values with defaults, with special
// presets for dev
func (c *IPFSOrchestratorConfig) SetDefaults(dev bool) {
	// API settings
	if c.API.Host == "" {
		c.API.Host = "127.0.0.1"
	}
	if c.API.Port == "" {
		c.API.Port = "9111"
	}
	if c.API.Key == "" {
		c.API.Key = "DO_NOT_LEAVE_ME_AS_DEFAULT"
	}

	// Proxy settings
	if c.Delegator.Host == "" {
		c.Delegator.Host = "127.0.0.1"
	}
	if c.Delegator.Port == "" {
		if dev {
			c.Delegator.Port = "8080"
		} else {
			c.Delegator.Port = "80"
		}
	}

	// Database settings
	if c.Database.URL == "" {
		c.Database.URL = "127.0.0.1"
	}
	if c.Database.Port == "" {
		if dev {
			c.Database.Port = "5433"
		} else {
			c.Database.Port = "5432"
		}
	}
	if dev {
		if c.Database.Username == "" {
			c.Database.Username = "postgres"
		}
		if c.Database.Password == "" {
			c.Database.Password = "password123"
		}
	}

	// IPFS settings
	if c.IPFS.Version == "" {
		c.IPFS.Version = DefaultIPFSVersion
	}
	if c.IPFS.DataDirectory == "" {
		if dev {
			c.IPFS.DataDirectory = "tmp"
		} else {
			c.IPFS.DataDirectory = "/"
		}
	}
	if c.IPFS.ModePerm == "" {
		c.IPFS.ModePerm = "0700"
	}
	if c.IPFS.Ports.Swarm == nil {
		c.IPFS.Ports.Swarm = []string{"4001-5000"}
	}
	if c.IPFS.Ports.API == nil {
		c.IPFS.Ports.API = []string{"5001-6000"}
	}
	if c.IPFS.Ports.Gateway == nil {
		c.IPFS.Ports.Gateway = []string{"8001-9000"}
	}
}

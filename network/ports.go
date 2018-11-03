package network

import (
	"errors"
	"net"

	"go.uber.org/zap"
)

// Registry manages host network usage
type Registry struct {
	l *zap.SugaredLogger

	host  string
	ports []string
}

// NewRegistry creates a new registry with given host address and available
// port ranges. Elements of portRanges can be "<PORT>" or "<LOWER>-<UPPER>"
func NewRegistry(logger *zap.SugaredLogger, host string, portRanges []string) *Registry {
	logger = logger.Named("network")
	if portRanges == nil {
		return &Registry{l: logger, host: host, ports: make([]string, 0)}
	}
	return &Registry{l: logger, host: host, ports: parsePorts(portRanges)}
}

// AssignPort assigns an available port and returns it
func (reg *Registry) AssignPort() (string, error) {
	for _, p := range reg.ports {
		l, err := net.Listen("tcp", reg.host+":"+p)
		if err != nil {
			continue
		}
		l.Close()
		return p, nil
	}
	return "", errors.New("no available port found")
}

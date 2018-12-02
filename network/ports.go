package network

import (
	"errors"
	"net"
	"time"

	"go.uber.org/zap"
)

// Registry manages host network usage
type Registry struct {
	l *zap.SugaredLogger

	host  string
	ports []string

	recent *cache
}

// NewRegistry creates a new registry with given host address and available
// port ranges. Elements of portRanges can be "<PORT>" or "<LOWER>-<UPPER>"
func NewRegistry(logger *zap.SugaredLogger, host string, portRanges []string) *Registry {
	var l = logger.Named("network")

	// mark available ports
	var ports []string
	if portRanges == nil {
		l.Warn("no port ranges were provided")
		ports = make([]string, 0)
	} else {
		ports = parsePorts(portRanges)
	}

	// set up cache
	c := newCache(5*time.Minute, 10*time.Minute)

	return &Registry{
		l:      l,
		host:   host,
		ports:  ports,
		recent: c,
	}
}

// AssignPort assigns an available port and returns it
func (reg *Registry) AssignPort() (string, error) {
	for reg.recent.Size() != len(reg.ports) {
		// roll random port to claim
		roll := random(len(reg.ports))
		p := reg.ports[roll]

		// if in cache, skip
		if found := reg.recent.Exists(p); found {
			continue
		}

		// attempt to claim port, placing it in cache
		reg.recent.Cache(p)
		l, err := net.Listen("tcp", reg.host+":"+p)
		if err != nil {
			continue
		}
		l.Close()

		return p, nil
	}

	// if loop exists, no port was found
	return "", errors.New("no available port found")
}

// Close stops the registry cache cleanup
func (reg *Registry) Close() {
	reg.recent.stop <- true
}

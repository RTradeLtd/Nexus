package network

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// Registry manages host network usage
type Registry struct {
	host  string
	ports []string
}

// NewRegistry creates a new registry with given host address and available
// port ranges. Elements of portRanges can be "<PORT>" or "<LOWER>-<UPPER>"
func NewRegistry(host string, portRanges []string) *Registry {
	if portRanges == nil {
		return &Registry{host: host, ports: make([]string, 0)}
	}
	return &Registry{host: host, ports: parsePorts(portRanges)}
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

func parsePorts(portRanges []string) []string {
	allPorts := make([]string, 0)
	for _, r := range portRanges {
		if strings.Contains(r, "-") {
			// is range
			ports := strings.Split(r, "-")
			lower, err := strconv.Atoi(ports[0])
			if err != nil {
				continue
			}
			upper, err := strconv.Atoi(ports[1])
			if err != nil {
				continue
			}
			for p := lower; p <= upper; p++ {
				allPorts = append(allPorts, strconv.Itoa(p))
			}
		} else {
			// check if int
			if _, err := strconv.Atoi(r); err != nil {
				continue
			}
			// is single port
			allPorts = append(allPorts, r)
		}
	}
	return allPorts
}

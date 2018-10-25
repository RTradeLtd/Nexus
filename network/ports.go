package network

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Registry manages host network usage
type Registry struct {
	ports map[string]net.Listener
	m     sync.Mutex
}

// NewRegistry creates a new registry with given host address and available
// port ranges. Elements of portRanges can be "<PORT>" or "<LOWER>-<UPPER>"
func NewRegistry(host string, portRanges []string) *Registry {
	reg := &Registry{ports: make(map[string]net.Listener)}
	if portRanges == nil {
		return reg
	}
	pts := parsePorts(portRanges)
	for _, p := range pts {
		// attempt to claim port
		reg.ports[p], _ = net.Listen("tcp", host+":"+p)
	}
	return reg
}

// AssignPort assigns an available port and returns it
func (reg *Registry) AssignPort() (string, error) {
	reg.m.Lock()
	for p, lock := range reg.ports {
		if lock != nil {
			lock.Close()
			reg.ports[p] = nil
			reg.m.Unlock()
			return p, nil
		}
	}
	reg.m.Unlock()
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

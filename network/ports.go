package network

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Registry manages host network usage
type Registry struct {
	host  string
	ports map[string]net.Listener
	m     sync.Mutex
}

// NewRegistry creates a new registry with given host address and available
// port ranges. Elements of portRanges can be "<PORT>" or "<LOWER>-<UPPER>"
func NewRegistry(host string, portRanges []string) *Registry {
	reg := &Registry{host: host, ports: make(map[string]net.Listener)}
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

// DeassignPort takes a previously assigned port and makes it available to
// Registry::AssignPort again
func (reg *Registry) DeassignPort(port string) error {
	if port == "" {
		return errors.New("invalid port")
	}

	reg.m.Lock()
	defer reg.m.Unlock()

	// check state of port lock in registry
	prevlock, found := reg.ports[port]
	if !found {
		return fmt.Errorf("port %s is not known to the registry", port)
	}
	if prevlock != nil {
		return fmt.Errorf("port %s is already available", port)
	}

	// lock the deassigned port and make available in registry
	lock, err := net.Listen("tcp", reg.host+":"+port)
	if err != nil {
		return fmt.Errorf("failed to claim port '%s:%s': %s", reg.host, port, err.Error())
	}
	reg.ports[port] = lock

	return nil
}

// Close releases all locked ports
func (reg *Registry) Close() {
	for port, lock := range reg.ports {
		if lock != nil {
			lock.Close()
			delete(reg.ports, port)
		}
	}
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

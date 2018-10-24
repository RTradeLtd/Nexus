package registry

import (
	"net"
	"strconv"
	"strings"
)

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

func lockPorts(host string, portRanges []string, reg map[string]net.Listener) {
	if portRanges == nil || reg == nil {
		return
	}
	pts := parsePorts(portRanges)
	for _, p := range pts {
		// attempt to claim port
		reg[p], _ = net.Listen("tcp", host+":"+p)
	}
}

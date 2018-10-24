package registry

import (
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
			// is single port
			allPorts = append(allPorts, r)
		}
	}
	return allPorts
}

package registry

import (
	"strings"
)

func parsePorts(portRanges []string) []string {
	allPorts := make([]string, 0)
	for _, r := range portRanges {
		if strings.Contains(r, "-") {
			// is range
			ports := strings.Split(r, "-")
			for _, p := range ports {
				allPorts = append(allPorts, p)
			}
		} else {
			// is single port
			allPorts = append(allPorts, r)
		}
	}
	return allPorts
}

package delegator

import "strings"

func stripLeadingSegments(path string) string {
	var expected = 5
	var parts = strings.SplitN(path, "/", expected)
	if len(parts) == expected {
		return "/" + parts[expected-1]
	}
	return path
}

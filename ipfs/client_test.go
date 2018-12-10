package ipfs

import (
	"strings"
	"testing"

	"github.com/RTradeLtd/ipfs-orchestrator/log"
)

func Test_client_getDataDir(t *testing.T) {
	l, _ := log.NewLogger("", true)
	var c = &client{l: l, dataDir: "./tmp", fileMode: 0755}
	d := c.getDataDir("path")
	if !strings.Contains(d, "path") {
		t.Errorf("expected 'path' in path, got %s", d)
	}
	if !strings.HasPrefix(d, "/") {
		t.Errorf("expected absolute path, got %s", d)
	}
}

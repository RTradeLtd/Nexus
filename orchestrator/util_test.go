package orchestrator

import "testing"

func Test_generateID(t *testing.T) {
	if id := generateID(); id == "" {
		t.Errorf("invalid ID generated")
	}
}

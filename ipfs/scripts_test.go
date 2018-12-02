package ipfs

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_newNodeStartScript(t *testing.T) {
	var (
		disk = 10
	)

	f, _ := ioutil.ReadFile("./internal/ipfs_start.sh")
	expected := fmt.Sprintf(string(f), disk)

	got, err := newNodeStartScript(disk)
	if err != nil {
		t.Error("unexpected err:", err.Error())
		return
	}
	if got != expected {
		t.Errorf("expected '%s', got '%s'", expected, got)
	}
}

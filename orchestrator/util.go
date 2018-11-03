package orchestrator

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func generateID() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}

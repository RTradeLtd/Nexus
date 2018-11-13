package network

import (
	"math/rand"
	"time"
)

func random(max int) int {
	if max <= 0 {
		return 0
	}
	rand.Seed(time.Now().Unix())
	return rand.Intn(max)
}

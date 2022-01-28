package trace

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return rand.Int63n(max-min) + min
}

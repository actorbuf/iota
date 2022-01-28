package trace

import (
	"fmt"
	"testing"
	"time"
)

func TestRandInt64(t *testing.T) {
	for i := 0; i < 20; i++ {
		go func() {
			fmt.Println(RandInt64(0, 2))
		}()
	}
	time.Sleep(time.Second)
}
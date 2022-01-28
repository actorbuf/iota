package trace

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var changLock sync.Mutex

func TestSamplerFuc(t *testing.T) {
	config := &SamplerStrategyConfig{
		Type:        SamplerTypeProbability,
		Probability: 0.3,
	}
	trueCount := 0
	falseCount := 0
	for i := 0; i < 10000; i++ {
		go func() {
			if _samplerFuc(config) {
				changLock.Lock()
				trueCount++
				changLock.Unlock()
			} else {
				changLock.Lock()
				falseCount++
				changLock.Unlock()
			}
		}()
	}
	time.Sleep(time.Second)
	fmt.Println("trueCount", trueCount)
	fmt.Println("falseCount", falseCount)
}

func TestRandInt64(t *testing.T) {
	for i := 0; i < 20; i++ {
		go func() {
			fmt.Println(RandInt64(0, 2))
		}()
	}
	time.Sleep(time.Second)
}
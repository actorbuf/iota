package utils

import (
	"fmt"
	"testing"
)

func TestGetTimeNowMillisecond(t *testing.T) {
	fmt.Println(GetTimeNowUnixMilli())
}

func TestUnixMilli2Time(t *testing.T) {
	fmt.Println(UnixMilli2Time(1631693848445).Unix())
}

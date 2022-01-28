package utils

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestPluck(t *testing.T) {
	type S struct {
		Id uint64
	}

	arr := []S{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
	}

	out := PluckUint64(arr, "Id")
	log.Infof("out %+v", out)

	out = PluckUint64("", "Id")
	log.Infof("out %+v", out)
}

func TestKeyBy(t *testing.T) {
	type T struct {
		Id   uint32
		Name string
	}

	list := []T{
		{
			Id:   1,
			Name: "hello",
		},
		{
			Id:   2,
			Name: "world",
		},
	}

	var m map[uint32]T
	KeyByMap(list, "Id", &m)

	log.Infof("m is %+v", m)
}

func TestChunk(t *testing.T) {
	var list = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 22, 22, 33, 4, 4, 3, 4, 234, 234, 23, 42, 34, 234, 23, 4, 234, 23, 423, 4, 234, 23, 4, 234, 23, 4, 234, 23, 4, 23, 4, 234, 23, 4, 423}
	res := Chunk(list, 10)
	// res = res.([][]int)
	fmt.Println(res)
}

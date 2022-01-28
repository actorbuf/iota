package utils

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

type A struct {
	I int
}

type B struct {
	AA *A
}

func TestReflectGetInt(t *testing.T) {
	var b B
	b.AA = &A{}
	b.AA.I = 10
	var h string
	i, success := ReflectGetInt(&b, "AA.I", &h)
	if success {
		log.Infof("got %d", i)
	} else {
		log.Warnf("get fail %s", h)
	}
}

func TestInterface2Int(t *testing.T) {
	i := Interface2Int(int64(-10))
	log.Infof("i %d", i)
}

func TestReflectGetStr(t *testing.T) {

	var typ = &struct {
		A string
		B uint64
	}{
		A: "hello",
		B: 100,
	}

	var c interface{} = typ

	res, _ := ReflectGetStr(c, "A", nil)
	log.Infof(res)
}

func TestClearSlicePtr(t *testing.T) {
	l := []int{1, 2, 3}
	p := &l
	ClearSlice(&p)
}

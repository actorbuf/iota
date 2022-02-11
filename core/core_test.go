package core

import (
	"fmt"
	"reflect"
	"testing"
)

func TestResponseCompatible(t *testing.T) {
	var err = &TestStruct{}

	data := ResponseCompatible(err)

	b, _ := json.Marshal(data)
	fmt.Println(string(b))
}

func TestName(t *testing.T) {
	var err interface{}

	var cerr = CreateErrorWithMsg(10, "你好")

	err = cerr

	fmt.Println(reflect.TypeOf(err).String() == "*core.ErrMsg")
}

package core

import (
	"fmt"
	"reflect"
	"testing"
)

type A struct {
}

func (a *A) RetErr() error {
	var err error
	return err
}

func TestRegister_getCallFunc(t *testing.T) {
	var a = new(A)
	fvo := reflect.ValueOf(a)
	ftp := fvo.Type()
	ff := ftp.Method(0).Func

	fret := ff.Call([]reflect.Value{fvo})
	fmt.Println(fret[0].Interface() == nil)
	//fmt.Println(fret[1].Interface())
}

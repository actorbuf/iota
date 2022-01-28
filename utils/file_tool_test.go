package utils

import (
	"fmt"
	"testing"
)

func TestGetFileLineOne(t *testing.T) {
	modFile := "./go.mod"
	exist := IsExist(modFile)
	if exist {
		text, err := GetFileLineOne(modFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(len(text[7:len(text) - 1]))
	} else {
		fmt.Println("go.mod is not exist! please run 'go mod init' ")
	}
}
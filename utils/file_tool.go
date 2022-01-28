package utils

import (
	"bufio"
	"errors"
	"io"
	"os"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func GetFileLineOne(filePath string) (lineOneText string, err error) {
	lineOneText = ""
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		err = errors.New(filePath + " open file error: " + err.Error())
		return
	}
	//建立缓冲区，把文件内容放到缓冲区中
	buf := bufio.NewReader(f)
	//遇到\n结束读取
	b, err := buf.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			err = errors.New(filePath + " is empty! ")
			return
		}
		err = errors.New(filePath + " read bytes error: " + err.Error())
		return
	}
	lineOneText = string(b)
	err = nil
	return
}

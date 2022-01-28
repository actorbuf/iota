package error

import (
	"errors"
	"fmt"
)

// NewLineCoverError 行重复覆盖错误
func NewLineCoverError(lineI int) error {
	text := fmt.Sprintf("line %v is cover", lineI)
	return &LineCoverErrorString{text}
}

type LineCoverErrorString struct {
	s string
}

func (e *LineCoverErrorString) Error() string {
	return e.s
}

// StructBuilderDataErr 结构体创建是否未输入结构体
var StructBuilderDataErr = errors.New("input data must be strut")

// StructBuilderFieldErr 结构体创建时对应的结构体不能转为string（当前支持类型看：fieldToString() ）
var StructBuilderFieldErr = errors.New("input field not to string")

package core

import (
	"reflect"
)

type Result struct {
	ErrCode int         `json:"err_code"`
	ErrMsg  string      `json:"err_msg"`
	Hint    string      `json:"hint,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func valueIsNil(vo reflect.Value) bool {
	if vo.Kind() == reflect.Ptr || vo.Kind() == reflect.Slice ||
		vo.Kind() == reflect.Map || vo.Kind() == reflect.Interface {
		if vo.IsNil() {
			return true
		}
	}
	return false
}

func ResponseCompatible(data interface{}) interface{} {
	// 为了兼容前端 不返回null值
	if data == nil {
		return []string{}
	}

	vo := reflect.ValueOf(data)
	if valueIsNil(vo) {
		data = []string{}
		return data
	}

	return data
}

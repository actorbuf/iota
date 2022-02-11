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

	// 是否兜一层
	//if vo.Kind() == reflect.Ptr {
	//	vo = vo.Elem()
	//}
	//if vo.Kind() == reflect.Struct {
	//	for i := 0; i < vo.NumField(); i++ {
	//		field := vo.Field(i)
	//		if valueIsNil(field) {
	//			if field.CanSet() {
	//				fieldType := field.Type()
	//				if fieldType.Kind() == reflect.Ptr {
	//					fieldType = fieldType.Elem()
	//				}
	//				realv := reflect.New(fieldType)
	//				field.Set(realv)
	//			}
	//		}
	//	}
	//}

	return data
}

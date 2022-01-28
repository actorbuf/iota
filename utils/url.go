package utils

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// Struct2UrlValues 将struct转为 url.Values 结构 只操作第一层 不嵌套
// 如果有json标签 将依照json标签作为key 否则按照字段名做key
// 如果struct字段为空 依然会将空值写入 url.Values 中
func Struct2UrlValues(obj interface{}) url.Values {
	if obj == nil {
		panic("object is nil")
	}

	var u = url.Values{}
	vo := reflect.ValueOf(obj)
	to := reflect.TypeOf(obj)

	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if to.Kind() == reflect.Ptr {
		to = to.Elem()
	}
	for i := 0; i < vo.NumField(); i++ {
		fieldType := to.Field(i)

		if fieldType.PkgPath != "" {
			continue
		}

		fieldName := fieldType.Name
		fieldTag := fieldType.Tag.Get("json")
		if fieldTag != "" {
			fieldName = strings.Replace(fieldTag, ",omitempty", "", 1)
		}
		fieldValue := vo.Field(i)
		if fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Struct ||
			fieldValue.Kind() == reflect.Map || fieldValue.Kind() == reflect.Slice ||
			fieldValue.Kind() == reflect.Array || fieldValue.Kind() == reflect.Chan ||
			fieldValue.Kind() == reflect.Func || fieldValue.Kind() == reflect.Interface ||
			fieldValue.Kind() == reflect.Complex64 || fieldValue.Kind() == reflect.Complex128 ||
			fieldValue.Kind() == reflect.Invalid || fieldValue.Kind() == reflect.Uintptr ||
			fieldValue.Kind() == reflect.UnsafePointer {
			continue
		}
		v := fmt.Sprintf("%+v", fieldValue.Interface())
		u.Set(fieldName, v)
	}
	return u
}

// Struct2UrlValuesOmitEmpty Struct2UrlValues 忽略空值字段
func Struct2UrlValuesOmitEmpty(obj interface{}) url.Values {
	if obj == nil {
		panic("object is nil")
	}

	var u = url.Values{}
	vo := reflect.ValueOf(obj)
	to := reflect.TypeOf(obj)

	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if to.Kind() == reflect.Ptr {
		to = to.Elem()
	}
	for i := 0; i < vo.NumField(); i++ {
		fieldType := to.Field(i)

		if fieldType.PkgPath != "" {
			continue
		}

		fieldName := fieldType.Name
		fieldTag := fieldType.Tag.Get("json")
		if fieldTag != "" {
			fieldName = strings.Replace(fieldTag, ",omitempty", "", 1)
		}
		fieldValue := vo.Field(i)
		if fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Struct ||
			fieldValue.Kind() == reflect.Map || fieldValue.Kind() == reflect.Slice ||
			fieldValue.Kind() == reflect.Array || fieldValue.Kind() == reflect.Chan ||
			fieldValue.Kind() == reflect.Func || fieldValue.Kind() == reflect.Interface ||
			fieldValue.Kind() == reflect.Complex64 || fieldValue.Kind() == reflect.Complex128 ||
			fieldValue.Kind() == reflect.Invalid || fieldValue.Kind() == reflect.Uintptr ||
			fieldValue.Kind() == reflect.UnsafePointer {
			continue
		}

		if fieldValue.IsZero() {
			continue
		}
		v := fmt.Sprintf("%+v", fieldValue.Interface())
		u.Set(fieldName, v)
	}
	return u
}

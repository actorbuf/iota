package util

import (
	"reflect"
	"strconv"
)

// NumToString 将数字转为string（TODO 存在float64转string的话可能会失真）
func NumToString(s interface{}) (bool, string) {
	field := reflect.ValueOf(s)
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true, strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true, strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return true, strconv.FormatFloat(field.Float(), 'f', -1, 64)
	default:
		return false, ""
	}
}

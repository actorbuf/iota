package utils

import (
	"errors"
	"fmt"
	"reflect"
)

// PluckUint64 从一个struct list 中 抽出某uint64字段的slice
func PluckUint64(list interface{}, fieldName string) []uint64 {

	var result []uint64

	vo := reflect.ValueOf(list)

	switch vo.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < vo.Len(); i++ {
			elem := vo.Index(i)
			for elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Struct {
				err := errors.New("element not struct")
				panic(err)
			}

			f := elem.FieldByName(fieldName)
			if !f.IsValid() {
				err := fmt.Errorf("struct missed field %s", fieldName)
				panic(err)
			}

			if f.Kind() != reflect.Uint64 {
				err := fmt.Errorf("struct element %s type required uint64", fieldName)
				panic(err)
			}

			result = append(result, f.Uint())
		}
	default:
		err := errors.New("required list of struct type")
		panic(err)
	}

	return result
}

// PluckUint64Map 从一个struct list 中 抽出某uint64字段的map
func PluckUint64Map(list interface{}, fieldName string) map[uint64]bool {
	out := PluckUint64(list, fieldName)
	res := map[uint64]bool{}
	for _, v := range out {
		res[v] = true
	}
	return res
}

func PluckUint32(list interface{}, fieldName string) []uint32 {
	var result []uint32

	vo := reflect.ValueOf(list)

	switch vo.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < vo.Len(); i++ {
			elem := vo.Index(i)
			for elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Struct {
				err := errors.New("element not struct")
				panic(err)
			}

			f := elem.FieldByName(fieldName)
			if !f.IsValid() {
				err := fmt.Errorf("struct missed field %s", fieldName)
				panic(err)
			}

			if f.Kind() != reflect.Uint32 {
				err := fmt.Errorf("struct element %s type required uint32", fieldName)
				panic(err)
			}

			result = append(result, uint32(f.Uint()))
		}
	default:
		err := errors.New("required list of struct type")
		panic(err)
	}

	return result
}

func PluckUint32Map(list interface{}, fieldName string) map[uint32]bool {
	out := PluckUint32(list, fieldName)
	res := map[uint32]bool{}
	for _, v := range out {
		res[v] = true
	}
	return res
}

// PluckString 从一个struct list 中 抽出某string字段的slice
func PluckString(list interface{}, fieldName string) []string {

	var result []string

	vo := reflect.ValueOf(list)

	switch vo.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < vo.Len(); i++ {
			elem := vo.Index(i)
			for elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Struct {
				err := errors.New("element not struct")
				panic(err)
			}

			f := elem.FieldByName(fieldName)
			if !f.IsValid() {
				err := fmt.Errorf("struct missed field %s", fieldName)
				panic(err)
			}

			if f.Kind() != reflect.String {
				err := fmt.Errorf("struct element %s type required string", fieldName)
				panic(err)
			}

			result = append(result, f.String())
		}
	default:
		err := errors.New("required list of struct type")
		panic(err)
	}

	return result
}

func PluckStringMap(list interface{}, fieldName string) map[string]bool {
	out := PluckString(list, fieldName)
	res := map[string]bool{}
	for _, v := range out {
		res[v] = true
	}
	return res
}

// KeyByMap 从一个struct list 中 根据某字段 转成一个以该字段为key的map
// list 是 []StructType
// res 是 *map[fieldType]StructType
func KeyByMap(list interface{}, fieldName string, res interface{}) {
	// 取下 field type
	vo := EnsureIsSliceOrArray(list)
	elType := vo.Type().Elem()

	t := elType
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("slice or array element required struct type, but got %v", t))
	}

	var keyType reflect.Type
	if sf, ok := t.FieldByName(fieldName); ok {
		keyType = sf.Type
	} else {
		panic(fmt.Sprintf("not found field %s", fieldName))
	}

	m := reflect.MakeMap(reflect.MapOf(keyType, elType))

	resVo := reflect.ValueOf(res)
	if resVo.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("invalid res type %v, required *map[key]val", resVo.Type()))
	}
	resVo = resVo.Elem()
	EnsureIsMapType(resVo, keyType, elType)

	l := vo.Len()
	for i := 0; i < l; i++ {
		el := vo.Index(i)
		elDef := el
		for elDef.Kind() == reflect.Ptr {
			elDef = elDef.Elem()
		}
		f := elDef.FieldByName(fieldName)
		if !f.IsValid() {
			continue
		}
		m.SetMapIndex(f, el)
	}

	resVo.Set(m)
}

// BaseTypeInArray 基本数据类型的判断是否在数组内，是则返回true以及下标
func BaseTypeInArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		length := s.Len()
		for i := 0; i < length; i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}
	return
}

// Chunk 按size分批 将 list => [size]sub-list-slice
func Chunk(list interface{}, size int) interface{} {
	vo := reflect.ValueOf(list)
	vt := reflect.TypeOf(list)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Slice && vo.Kind() != reflect.Array {
		panic("chunk first argument must be a slice")
	}
	length := vo.Len()
	capGrow := length % size
	if capGrow > 0 {
		capGrow = 1
	}
	ss := reflect.MakeSlice(reflect.SliceOf(vt), 0, length/size+capGrow)

	var start, end int
	for start < length {
		end = start + size
		if end > length {
			end = length
		}
		ss = reflect.Append(ss, vo.Slice(start, end))
		start = end
	}

	return ss.Interface()
}

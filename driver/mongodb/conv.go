package mongodb

import (
	"fmt"
	"reflect"
	"strings"
)

// 这里定义了一个特殊功能 格式转换

// Struct2MapWithBsonTag 结构体转map 用 bson字段名 做Key
// 注意 obj非struct结构将PANIC
func Struct2MapWithBsonTag(obj interface{}) map[string]interface{} {
	vo := reflect.ValueOf(obj)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Struct {
		panic("object type not struct")
	}
	var data = make(map[string]interface{})
	for i := 0; i < vo.NumField(); i++ {
		vf := vo.Field(i)
		key := vo.Type().Field(i).Tag.Get("bson")
		if key == "" {
			key = vo.Type().Field(i).Name
		}
		if vf.CanSet() {
			data[key] = vf.Interface()
		}
	}
	return data
}

// Struct2MapWithJsonTag 结构体转map 用 json字段名 做Key
// 注意 obj非struct结构将PANIC
func Struct2MapWithJsonTag(obj interface{}) map[string]interface{} {
	vo := reflect.ValueOf(obj)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Struct {
		panic("object type not struct")
	}
	var data = make(map[string]interface{})
	for i := 0; i < vo.NumField(); i++ {
		vf := vo.Field(i)
		key := vo.Type().Field(i).Tag.Get("json")
		if key == "" {
			key = vo.Type().Field(i).Name
		}
		// 过滤掉 omitempty 选项
		if strings.Contains(key, ",omitempty") {
			key = strings.Replace(key, ",omitempty", "", 1)
		}
		if vf.CanSet() {
			data[key] = vf.Interface()
		}
	}
	return data
}

// Struct2MapOmitEmpty 结构体转map并忽略空字段 用 字段名 做Key
// 注意 只忽略顶层字段 obj非struct结构将PANIC
func Struct2MapOmitEmpty(obj interface{}) map[string]interface{} {
	vo := reflect.ValueOf(obj)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Struct {
		panic("object type not struct")
	}
	var data = make(map[string]interface{})
	for i := 0; i < vo.NumField(); i++ {
		vf := vo.Field(i)
		if !vf.IsZero() && vf.CanSet() {
			data[vo.Type().Field(i).Name] = vf.Interface()
		}
	}
	return data
}

// Struct2MapOmitEmptyWithBsonTag 结构体转map并忽略空字段 按照 bson 标签做key
// obj 需要是一个指针
// 注意 只忽略顶层字段 obj非struct结构将PANIC
func Struct2MapOmitEmptyWithBsonTag(obj interface{}) map[string]interface{} {
	vo := reflect.ValueOf(obj)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Struct {
		panic("object type not struct")
	}
	var data = make(map[string]interface{})
	for i := 0; i < vo.NumField(); i++ {
		vf := vo.Field(i)
		key := vo.Type().Field(i).Tag.Get("bson")
		if key == "" {
			key = vo.Type().Field(i).Name
		}
		if !vf.IsZero() && vf.CanSet() {
			data[key] = vf.Interface()
		}
	}
	return data
}

// Struct2MapOmitEmptyWithJsonTag 结构体转map并忽略空字段 按照 json 标签做key
// 注意 只忽略顶层字段 obj非struct结构将PANIC
func Struct2MapOmitEmptyWithJsonTag(obj interface{}) map[string]interface{} {
	vo := reflect.ValueOf(obj)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Struct {
		panic("object type not struct")
	}
	var data = make(map[string]interface{})
	for i := 0; i < vo.NumField(); i++ {
		vf := vo.Field(i)
		key := vo.Type().Field(i).Tag.Get("json")
		if key == "" {
			key = vo.Type().Field(i).Name
		}
		// 过滤掉 omitempty 选项
		if strings.Contains(key, ",omitempty") {
			key = strings.Replace(key, ",omitempty", "", 1)
		}
		if !vf.IsZero() && vf.CanSet() {
			data[key] = vf.Interface()
		}
	}
	return data
}

// SliceStruct2MapOmitEmpty 结构体数组转map数组并忽略空字段
// 注意 子元素非struct将被忽略
func SliceStruct2MapOmitEmpty(obj interface{}) interface{} {
	vo := reflect.ValueOf(obj)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Slice && vo.Kind() != reflect.Array {
		panic("object type not slice")
	}
	var data []map[string]interface{}

	for i := 0; i < vo.Len(); i++ {
		node := vo.Index(i)
		if node.Kind() == reflect.Ptr {
			node = node.Elem()
		}
		if node.Kind() != reflect.Struct && node.Kind() != reflect.Interface {
			continue
		}
		fn := Struct2MapOmitEmpty(node.Interface())
		data = append(data, fn)
	}
	return data
}

// SetMapOmitInsertField 对于一个 update-bson-map 忽略$set中的$setOnInsert字段
// 需要传递一个map的指针 确保数据可写 否则PANIC
// $set支持map和struct 其他结构体将PANIC
// $setOnInsert只支持map 其他结构体将PANIC
// 只对包含$set和$setOnInsert的map生效 若$set或者$setOnInsert缺失 将PANIC
func SetMapOmitInsertField(m interface{}) {
	vo := reflect.ValueOf(m)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if vo.Kind() != reflect.Map {
		panic("object type not map")
	}
	setVal := vo.MapIndex(reflect.ValueOf("$set"))
	if !setVal.IsValid() {
		panic("$set not found")
	}
	soiVal := vo.MapIndex(reflect.ValueOf("$setOnInsert"))
	if !soiVal.IsValid() {
		panic("$setOnInsert not found")
	}
	if !vo.CanSet() {
		panic("map can't set")
	}
	soiRealVal := reflect.ValueOf(soiVal.Interface())
	if soiRealVal.Kind() == reflect.Ptr {
		soiRealVal = soiRealVal.Elem()
	}
	if soiRealVal.Kind() != reflect.Map {
		err := fmt.Errorf("$setOnInsert type not map: type(%v)", soiRealVal.Kind())
		panic(err)
	}
	setRealTyp := reflect.TypeOf(setVal.Interface())
	if setRealTyp.Kind() == reflect.Ptr {
		setRealTyp = setRealTyp.Elem()
	}
	setRealVal := reflect.ValueOf(setVal.Interface())
	if setRealVal.Kind() == reflect.Ptr {
		setRealVal = setRealVal.Elem()
	}
	if setRealVal.Kind() != reflect.Struct && setRealVal.Kind() != reflect.Map {
		err := fmt.Errorf("$set type not map or struct: type(%v)", setRealVal.Kind())
		panic(err)
	}
	var setMap = make(map[string]interface{})

	//builder := RegisterTimestampCodec(nil).Build()
	if setRealVal.Kind() == reflect.Struct {
		var data = make(map[string]interface{})
		for i := 0; i < setRealVal.NumField(); i++ {
			field := setRealTyp.Field(i)
			bsonTag := field.Tag.Get("bson")
			if bsonTag == "" {
				continue
			}
			fieldVal := setRealVal.Field(i)
			data[bsonTag] = fieldVal.Interface()
		}
		vo.SetMapIndex(reflect.ValueOf("$set"), reflect.ValueOf(data))
	}

	setRealVal = reflect.ValueOf(vo.MapIndex(reflect.ValueOf("$set")).Interface())

	if setRealVal.Kind() == reflect.Ptr {
		setRealVal = setRealVal.Elem()
	}

	if setRealVal.Kind() == reflect.Map {
		iter := setRealVal.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()
			soiFType := soiRealVal.MapIndex(reflect.ValueOf(key.String()))
			if !soiFType.IsValid() {
				setMap[key.String()] = val.Interface()
			}
		}
	}
	vo.SetMapIndex(reflect.ValueOf("$set"), reflect.ValueOf(setMap))
}

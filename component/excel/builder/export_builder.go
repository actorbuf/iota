package builder

import (
	error2 "github.com/actorbuf/iota/component/excel/error"
	"reflect"
	"strings"
)

const ExcelHeadTag = "excel_head"
const ExcelExclude = "excel_exclude"
const ExcelExcludeTrue = "true"

// DataBuilder 数据来源builder
type DataBuilder interface {
	GetHeads() [][]interface{}                // 获取头部数据
	GetLines() (map[int][]interface{}, error) // 获取行数据
}

var _ DataBuilder = new(ArrDataBuilder)
var _ DataBuilder = new(StructDataBuilder)

// ArrDataBuilder 通过二维数组创建数据
type ArrDataBuilder struct {
	heads [][]interface{}
	lines map[int][]interface{}
}

// AddHead 添加头信息
func (dataBuilder *ArrDataBuilder) AddHead(head []interface{}) *ArrDataBuilder {
	dataBuilder.heads = append(dataBuilder.heads, head)
	return dataBuilder
}

// AddHeads 添加头信息
func (dataBuilder *ArrDataBuilder) AddHeads(heads [][]interface{}) *ArrDataBuilder {
	dataBuilder.heads = heads
	return dataBuilder
}

// AddLine 添加行信息； line 第几行
func (dataBuilder *ArrDataBuilder) AddLine(line int, data []interface{}) *ArrDataBuilder {
	if dataBuilder.lines == nil {
		dataBuilder.lines = make(map[int][]interface{})
	}
	dataBuilder.lines[line] = data
	return dataBuilder
}

// AddLines 添加行信息；
func (dataBuilder *ArrDataBuilder) AddLines(lines map[int][]interface{}) *ArrDataBuilder {
	dataBuilder.lines = lines
	return dataBuilder
}

func (dataBuilder *ArrDataBuilder) GetHeads() [][]interface{} {
	return dataBuilder.heads
}

func (dataBuilder *ArrDataBuilder) GetLines() (map[int][]interface{}, error) {
	return dataBuilder.lines, nil
}

// StructDataBuilder 通过结构体建立数据（当前结构体不支持变长）
type StructDataBuilder struct {
	heads [][]interface{}
	lines map[int]interface{}
}

// AddHead 添加头信息
func (dataBuilder *StructDataBuilder) AddHead(head []interface{}) *StructDataBuilder {
	dataBuilder.heads = append(dataBuilder.heads, head)
	return dataBuilder
}

// AddHeadByStruct 通过结构体添加头信息
func (dataBuilder *StructDataBuilder) AddHeadByStruct(s interface{}) *StructDataBuilder {
	value := reflect.TypeOf(s)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return dataBuilder
	}
	head := make([]interface{}, 0)
	for i := 0; i < value.NumField(); i++ {
		tag := value.Field(i).Tag
		// 去掉不需要excel处理的字段
		exclude := tag.Get(ExcelExclude)
		if exclude != "" && strings.Contains(exclude, ExcelExcludeTrue) {
			continue
		}
		tagHead := tag.Get(ExcelHeadTag)
		head = append(head, tagHead)
	}
	if len(head) > 0 {
		dataBuilder.AddHead(head)
	}
	return dataBuilder
}

// AddHeads 添加头信息
func (dataBuilder *StructDataBuilder) AddHeads(heads [][]interface{}) *StructDataBuilder {
	dataBuilder.heads = heads
	return dataBuilder
}

// AddLine 添加行信息； line 第几行
func (dataBuilder *StructDataBuilder) AddLine(line int, data interface{}) *StructDataBuilder {
	if dataBuilder.lines == nil {
		dataBuilder.lines = make(map[int]interface{})
	}
	dataBuilder.lines[line] = data
	return dataBuilder
}

// AddLines 添加行信息；
func (dataBuilder *StructDataBuilder) AddLines(lines map[int]interface{}) *StructDataBuilder {
	dataBuilder.lines = lines
	return dataBuilder
}

// AddStructs 按struct进行添加
func (dataBuilder *StructDataBuilder) AddStructs(lines []interface{}) *StructDataBuilder {
	// 转为map
	index := len(dataBuilder.heads)
	for i := range lines {
		dataBuilder.AddLine(i+index+1, lines[i])
	}
	return dataBuilder
}

// AddStructAndHead 按struct进行添加；同时拿Struct的Tag获取对应head
func (dataBuilder *StructDataBuilder) AddStructAndHead(lines []interface{}) *StructDataBuilder {
	// 按tag拿到对应head
	for i := range lines {
		value := reflect.TypeOf(lines[i])
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			continue
		}
		head := make([]interface{}, 0)
		for i := 0; i < value.NumField(); i++ {
			tag := value.Field(i).Tag
			// 去掉不需要excel处理的字段
			exclude := tag.Get(ExcelExclude)
			if exclude != "" && strings.Contains(exclude, ExcelExcludeTrue) {
				continue
			}
			tagHead := tag.Get(ExcelHeadTag)
			head = append(head, tagHead)
		}
		if len(head) > 0 {
			dataBuilder.AddHead(head)
		}
		break
	}

	// 转为map
	index := len(dataBuilder.heads)
	for i := range lines {
		dataBuilder.AddLine(i+index+1, lines[i])
	}
	return dataBuilder
}

func (dataBuilder *StructDataBuilder) GetHeads() [][]interface{} {
	return dataBuilder.heads
}

func (dataBuilder *StructDataBuilder) GetLines() (map[int][]interface{}, error) {
	m := make(map[int][]interface{}, len(dataBuilder.lines))
	for lineI := range dataBuilder.lines {
		value := reflect.ValueOf(dataBuilder.lines[lineI])
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			return nil, error2.StructBuilderDataErr
		}
		line := make([]interface{}, 0, value.NumField())
		valueT := value.Type()

		for i := 0; i < value.NumField(); i++ {
			tag := valueT.Field(i).Tag
			// 去掉不需要excel处理的字段
			exclude := tag.Get(ExcelExclude)
			if exclude != "" && strings.Contains(exclude, ExcelExcludeTrue) {
				continue
			}
			line = append(line, value.Field(i).Interface())
		}
		m[lineI] = line
	}

	return m, nil
}

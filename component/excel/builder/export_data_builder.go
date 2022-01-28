package builder

import (
	"github.com/actorbuf/iota/component/excel/util"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// ExportData 旧版本的数据导出
type ExportData struct {
	Sheet string     `json:"sheet"`
	Head  []string   `json:"heads"`
	Data  [][]string `json:"data"`
}

func NewData() *ExportData {
	return &ExportData{
		Sheet: "Sheet1",
		Head:  []string{},
		Data:  [][]string{},
	}
}

func (e *ExportData) SetSheet(sheet string) *ExportData {
	e.Sheet = sheet
	return e
}

func (e *ExportData) AddHead(head string) *ExportData {
	e.Head = append(e.Head, head)
	return e
}

func (e *ExportData) AddHeads(heads []string) *ExportData {
	e.Head = heads
	return e
}

func (e *ExportData) AddLine(line []string) *ExportData {
	e.Data = append(e.Data, line)
	return e
}

func (e *ExportData) AddLines(lines [][]string) *ExportData {
	e.Data = lines
	return e
}

func (e *ExportData) Export() (*excelize.File, error) {
	var err error
	f := excelize.NewFile()           // 创建文件
	sheetIndex := f.NewSheet(e.Sheet) // 创建工作簿
	f.SetActiveSheet(sheetIndex)      // 设置激活的工作簿
	var lineI = 1
	if len(e.Head) > 0 {
		// 先填充头
		for index, headData := range e.Head {
			lineIStr := strconv.FormatInt(int64(lineI), 10)
			err = f.SetCellStr(e.Sheet, util.ToLine(index+1)+lineIStr, headData)
			if err != nil {
				return f, err
			}
		}
		lineI++
	}
	for index := range e.Data {
		for lineIndex := range e.Data[index] {
			lineIStr := strconv.FormatInt(int64(lineI), 10)
			err = f.SetCellStr(e.Sheet, util.ToLine(lineIndex+1)+lineIStr, e.Data[index][lineIndex])
			if err != nil {
				return f, err
			}
		}
		lineI++
	}
	return f, err
}

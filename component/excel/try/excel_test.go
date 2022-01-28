package excel_test

import (
	"fmt"
	builder2 "github.com/actorbuf/iota/component/excel/builder"
	util2 "github.com/actorbuf/iota/component/excel/util"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

func TestLineNum(t *testing.T) {
	fmt.Println(util2.ToLine(52))
}

func TestToChar(t *testing.T) {
	fmt.Println(util2.ToLine(1))
}

func TestExportArr(t *testing.T) {
	filename := fmt.Sprintf("excel_%s.xlsx", time.Now().Format("20060102150405"))
	fd, err := os.Create(filename)
	defer func() {
		_ = fd.Close()
	}()
	dataBuilder := new(builder2.ArrDataBuilder)
	dataBuilder.AddHead([]interface{}{"aaaa", "bbbb", "cccc", "dddd"}).
		AddHead([]interface{}{"aaaa2", "bbbb2", "cccc2", "dddd2"})

	for i := 0; i < 10; i++ {
		line := make([]interface{}, 0, 8)
		for j := 0; j < 600; j++ {
			line = append(line, fmt.Sprintf("line %v char %v", i, j))
		}
		dataBuilder.AddLine(i+3, line)
	}

	sheet := builder2.NewSheet().SetSheetName("Sheet1").SetDataBuilder(dataBuilder)
	err = builder2.NewFile().AddSheet(sheet).ExportFile(fd)
	if err != nil {
		fmt.Println(err)
	}
}

func TestExportStrut(t *testing.T) {
	filename := fmt.Sprintf("excel_%s.xlsx", time.Now().Format("20060102150405"))
	fd, err := os.Create(filename)
	defer func() {
		_ = fd.Close()
	}()
	dataBuilder := new(builder2.StructDataBuilder)
	dataBuilder.AddHead([]interface{}{"ID", "Name", "Age", "State", "RobotWxIds"}).
		AddHead([]interface{}{"aaaa2", "bbbb2", "cccc2", "dddd2"})

	for i := 0; i < 10; i++ {
		line := struct {
			ID         string
			Name       string
			Age        int64
			State      int
			RobotWxIds []string `json:"robot_wx_ids"`
		}{
			ID:         fmt.Sprintf("ID %v ", i),
			Name:       fmt.Sprintf("Name %v ", i),
			Age:        int64(i),
			State:      i,
			RobotWxIds: []string{fmt.Sprintf("RobotWxID1 %v ", i), fmt.Sprintf("RobotWxID2 %v ", i)},
		}
		dataBuilder.AddLine(i+3, line)
	}

	sheet := builder2.NewSheet().SetDataBuilder(dataBuilder)
	err = builder2.NewFile().AddSheet(sheet).ExportFile(fd)
	if err != nil {
		fmt.Println(err)
	}
}

type Bson struct {
	Count          int     `json:"count" bson:"count" excel_head:"请求次数" excel_exclude:"true"`
	Requesting     int32   `json:"requesting" bson:"requesting" excel_head:"请求中"`
	RequestSuccess uint    `json:"request_success" bson:"request_success" excel_head:"请求成功"`
	RequestFail    uint32  `json:"request_fail" bson:"request_fail" excel_head:"请求失败"`
	AddWait        float64 `json:"add_wait" bson:"add_wait" excel_head:"等待通过"`
	AddSuccess     int64   `json:"add_success" bson:"add_success" excel_head:"通过好友数"`
	Dt             string  `json:"dt" bson:"dt" excel_head:"日期"`
}

func TestExportMongo2(t *testing.T) {
	filename := fmt.Sprintf("./excel_%s.xlsx", time.Now().Format("20060102150405"))
	fd, err := os.Create(filename)
	defer func() {
		_ = fd.Close()
	}()
	var arr []*Bson
	for i := 0; i < 10; i++ {
		arr = append(arr, &Bson{
			Count:          i,
			Requesting:     int32(i),
			RequestSuccess: uint(i),
			RequestFail:    uint32(i),
			AddWait:        float64(i),
			AddSuccess:     int64(i),
			Dt:             fmt.Sprintf("aaaa%v", i)})
	}

	arrI := make([]interface{}, 0, len(arr))
	//for i := range arr {
	//	arr[i].AddWait = 1232132311.01202122
	//	arr[i].AddSuccess = 123213231231232011
	//	arrI = append(arrI, arr[i])
	//}
	dataBuilder := new(builder2.StructDataBuilder)

	if len(arrI) == 0 {
		dataBuilder.AddHeadByStruct(new(Bson))
	} else {
		dataBuilder.AddStructAndHead(arrI)
	}

	sheet := builder2.NewSheet().SetDataBuilder(dataBuilder).
		AddColHeight(builder2.NewColHeightByNum(3, 100)).
		AddColWidth(builder2.NewColWidthByNum(2, 3, 50)).
		AddMergeCell(builder2.NewMergeCellByNum(1, 2, 2, 2)).
		SetNumKeep(true)

	err = builder2.NewFile().AddSheet(sheet).ExportFile(fd)
	if err != nil {
		fmt.Printf("error: %v, stack: %v \n", err, string(debug.Stack()))
	}
}

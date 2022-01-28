package builder

import (
	"fmt"
	util2 "github.com/actorbuf/iota/component/excel/util"
	"strconv"

	"github.com/xuri/excelize/v2"
)

var DefaultHeadStyle = excelize.Style{
	Font: &excelize.Font{
		Bold: true,
	},
	Alignment: &excelize.Alignment{
		Horizontal: "center",
		Vertical:   "center",
	},
}

// CellNum 单元格（数字方式；如：{ Column： 1， Line：1} 对应：A1）
type CellNum struct {
	Column int // 第几列（从1开始算）
	Line   int // 第几行（从1开始算）
}

// CellChar 单元格（字母方式;如：{ Cell: "A1"}）
type CellChar struct {
	Cell string
}

// Cell 单元格（得到对应的单元格信息）
type Cell struct {
	CellNum  *CellNum
	CellChar *CellChar
}

// GetCell 获取单元格的string信息
func (cell *Cell) GetCell() string {
	if cell.CellNum != nil {
		return util2.ToLine(cell.CellNum.Column) + strconv.FormatInt(int64(cell.CellNum.Line), 10)
	}
	if cell.CellChar != nil {
		return cell.CellChar.Cell
	}
	return ""
}

// CellArea 单元格区域
type CellArea struct {
	HCell *Cell
	VCell *Cell
}

func (area *CellArea) GetHCell() string {
	if area.HCell != nil {
		return area.HCell.GetCell()
	}
	return ""
}

func (area *CellArea) GetVCell() string {
	if area.VCell != nil {
		return area.VCell.GetCell()
	}
	return ""
}

// Style 改良版的Style
type Style struct {
	CellArea *CellArea
	Style    *excelize.Style
}

// NewStyleByNum 新建样式通过数字行列；如：(hColumn = 1, hLine = 1, vCollum = 3, vLine = 4) 对应"A1:C4"
func NewStyleByNum(hColumn, hLine, vCollum, vLine int, style *excelize.Style) *Style {
	return &Style{
		CellArea: &CellArea{
			HCell: &Cell{CellNum: &CellNum{Column: hColumn, Line: hLine}},
			VCell: &Cell{CellNum: &CellNum{Column: vCollum, Line: vLine}},
		},
		Style: style,
	}
}

// NewStyleByChar 新建样式通过字母；如：(hCell = "A1", vCell = "C4") 对应"A1:C4"
func NewStyleByChar(hCell, vCell string, style *excelize.Style) *Style {
	return &Style{
		CellArea: &CellArea{
			HCell: &Cell{CellChar: &CellChar{Cell: hCell}},
			VCell: &Cell{CellChar: &CellChar{Cell: vCell}},
		},
		Style: style,
	}
}

func (style *Style) GetHCell() string {
	return style.CellArea.GetHCell()
}

func (style *Style) GetVCell() string {
	return style.CellArea.GetVCell()
}

// GetCell 获取style对应的单元格；如："A1:C4"
func (style *Style) GetCell() string {
	hCell := "A1"
	vCell := "A1"
	if style.GetHCell() != "" {
		hCell = style.GetHCell()
	}
	if style.GetVCell() != "" {
		vCell = style.GetVCell()
	}

	return fmt.Sprintf("%s:%s", hCell, vCell)
}

// setStyle 设置样式
func setStyle(f *excelize.File, sheetName string, style *Style) error {
	styleID, err := f.NewStyle(style.Style)
	if err != nil {
		return err
	}
	return f.SetCellStyle(sheetName, style.GetHCell(), style.GetVCell(), styleID)
}

// MergeCell 合并单元格
type MergeCell struct {
	CellArea *CellArea
}

func (merge *MergeCell) GetHCell() string {
	return merge.CellArea.GetHCell()
}

func (merge *MergeCell) GetVCell() string {
	return merge.CellArea.GetVCell()
}

// NewMergeCellByNum 新建合并通过数字行列；如：(hColumn = 1, hLine = 1, vCollum = 3, vLine = 4) 对应"A1:C4"
func NewMergeCellByNum(hColumn, hLine, vCollum, vLine int) *MergeCell {
	return &MergeCell{
		CellArea: &CellArea{
			HCell: &Cell{CellNum: &CellNum{Column: hColumn, Line: hLine}},
			VCell: &Cell{CellNum: &CellNum{Column: vCollum, Line: vLine}},
		},
	}
}

// NewMergeCellByChar 新建合并通过字母；如：(hCell = "A1", vCell = "C4") 对应"A1:C4"
func NewMergeCellByChar(hCell, vCell string) *MergeCell {
	return &MergeCell{
		CellArea: &CellArea{
			HCell: &Cell{CellChar: &CellChar{Cell: hCell}},
			VCell: &Cell{CellChar: &CellChar{Cell: vCell}},
		},
	}
}

// mergeCell 合并单元格
func mergeCell(f *excelize.File, sheetName string, merge *MergeCell) error {
	return f.MergeCell(sheetName, merge.GetHCell(), merge.GetVCell())
}

// ColWidth 单元格宽度
type ColWidth struct {
	StartColumn     int    // 开始第几列（从1开始算）
	StartColumnChar string // 开始第几列（从"A"开始算）
	EndColumn       int    // 结束第几列（从1开始算）
	EndColumnChar   string // 结束第几列（从"A"开始算）
	Width           float64
}

func (colWidth *ColWidth) GetHCell() string {
	if colWidth.StartColumn > 0 {
		return util2.ToLine(colWidth.StartColumn)
	}

	return colWidth.StartColumnChar
}

func (colWidth *ColWidth) GetVCell() string {
	if colWidth.EndColumn > 0 {
		return util2.ToLine(colWidth.EndColumn)
	}

	return colWidth.EndColumnChar
}

// NewColWidthByNum 新建单元格宽度通过数字行列；如：(startCol = 1, endCol = 3) 对应"A~C"
func NewColWidthByNum(startCol, endCol int, width float64) *ColWidth {
	return &ColWidth{
		StartColumn: startCol,
		EndColumn:   endCol,
		Width:       width,
	}
}

// NewColWidthByChar 新建单元格宽度通过字母；如：(startCol = "A", endCol = "C") 对应"A~C"
func NewColWidthByChar(startCol, endCol string, width float64) *ColWidth {
	return &ColWidth{
		StartColumnChar: startCol,
		EndColumnChar:   endCol,
		Width:           width,
	}
}

// setColWidth 设置宽度
func setColWidth(f *excelize.File, sheetName string, colWidth *ColWidth) error {
	return f.SetColWidth(sheetName, colWidth.GetHCell(), colWidth.GetVCell(), colWidth.Width)
}

// ColHeight 单元格高度
type ColHeight struct {
	Line   int
	Height float64
}

// NewColHeightByNum 新建单元格高度通过数字行列；如：(hColumn = 1, hLine = 1, vCollum = 3, vLine = 4) 对应"A1:C4"
func NewColHeightByNum(line int, height float64) *ColHeight {
	return &ColHeight{
		Line:   line,
		Height: height,
	}
}

// setColWidth 设置高度
func setColHeight(f *excelize.File, sheetName string, colHeight *ColHeight) error {
	return f.SetRowHeight(sheetName, colHeight.Line, colHeight.Height)
}

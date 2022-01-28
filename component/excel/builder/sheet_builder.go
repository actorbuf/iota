package builder

import (
	error2 "github.com/actorbuf/iota/component/excel/error"
	"github.com/actorbuf/iota/component/excel/util"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type sheet struct {
	file         *File // excel文件
	sheetName    string
	heads        [][]interface{}       // 头信息
	lines        map[int][]interface{} // 行信息
	maxLine      int                   // 最大行
	maxChar      int                   // 最大列
	lineMap      map[int]struct{}      // 行是否被数据填充过，防止数据覆盖
	isStyleCover bool                  // 是否样式覆盖
	styleMap     map[string]struct{}   // cell是不是已经被修饰过，如果是，不再修饰 (TODO)
	DataBuilder  DataBuilder           // 数据构建来源
	headStyle    *excelize.Style       // 头部样式
	bodyStyle    []*Style              // 整体样式
	mergeCell    []*MergeCell          // 合并单元格
	colWidth     []*ColWidth           // 设置单元格宽度
	colHeight    []*ColHeight          // 设置单元格高度
	numKeep      bool                  // 是否保持为
}

// NewSheet 新建一个Sheet
func NewSheet(options ...SheetOptionFunc) *sheet {
	sheet := new(sheet)
	for i := range options {
		options[i](sheet)
	}
	return sheet
}

// NewSheetByOptions 新建一个Sheet
func NewSheetByOptions(options *SheetOptions) *sheet {
	sheet := new(sheet)
	for i := range options.options {
		options.options[i](sheet)
	}
	return sheet
}

// SheetOptionFunc SheetOptions sheet新建修饰方法
type SheetOptionFunc func(sheet *sheet)
type SheetOptions struct {
	options []SheetOptionFunc
}

func (options *SheetOptions) Append(option SheetOptionFunc) *SheetOptions {
	options.options = append(options.options, option)
	return options
}

// SetSheetName 设置sheet的名字
func SetSheetName(sheetName string) SheetOptionFunc {
	return func(sheet *sheet) {
		sheet.sheetName = sheetName
	}
}

// defaultDataBuilder 默认用二维数组创建数据
var defaultDataBuilder = new(ArrDataBuilder)

// SetDataBuilder 设置数据来源
func SetDataBuilder(dataBuilder DataBuilder) SheetOptionFunc {
	return func(sheet *sheet) {
		sheet.DataBuilder = dataBuilder
	}
}

// SetFile 设置文件
func SetFile(file *File) SheetOptionFunc {
	return func(sheet *sheet) {
		sheet.file = file
	}
}

// SetSheetName 设置sheet名
func (sheet *sheet) SetSheetName(sheetName string) *sheet {
	sheet.sheetName = sheetName
	return sheet
}

// SetDataBuilder 设置默认来源
func (sheet *sheet) SetDataBuilder(dataBuilder DataBuilder) *sheet {
	sheet.DataBuilder = dataBuilder
	return sheet
}

// SetFile 设置文件
func (sheet *sheet) SetFile(file *File) *sheet {
	sheet.file = file
	return sheet
}

// SetHeadStyle 设置头部样式
func (sheet *sheet) SetHeadStyle(headStyle *excelize.Style) *sheet {
	sheet.headStyle = headStyle
	return sheet
}

// SetBodyStyle 设置身体样式
func (sheet *sheet) SetBodyStyle(bodyStyle []*Style) *sheet {
	sheet.bodyStyle = bodyStyle
	return sheet
}

// AddBodyStyle 设置身体样式
func (sheet *sheet) AddBodyStyle(bodyStyle *Style) *sheet {
	sheet.bodyStyle = append(sheet.bodyStyle, bodyStyle)
	return sheet
}

// AddMergeCell 设置合并单元格
func (sheet *sheet) AddMergeCell(mergeCell *MergeCell) *sheet {
	sheet.mergeCell = append(sheet.mergeCell, mergeCell)
	return sheet
}

// AddColWidth 设置单元格宽
func (sheet *sheet) AddColWidth(colWidth *ColWidth) *sheet {
	sheet.colWidth = append(sheet.colWidth, colWidth)
	return sheet
}

// SetNumKeep 设置是否保持数字
func (sheet *sheet) SetNumKeep(keep bool) *sheet {
	sheet.numKeep = keep
	return sheet
}

// AddColHeight 设置单元格高
func (sheet *sheet) AddColHeight(colHeight *ColHeight) *sheet {
	sheet.colHeight = append(sheet.colHeight, colHeight)
	return sheet
}

// TODO 默认的一些格式化操作
// incMaxLine 增加最大行数
func (sheet *sheet) incMaxLine(lineI int) {
	if lineI > sheet.maxLine {
		sheet.maxLine = lineI
	}
}

// incMaxChar 增加最大列数
func (sheet *sheet) incMaxChar(charI int) {
	if charI > sheet.maxChar {
		sheet.maxChar = charI
	}
}

// lineUse 行被数据填充过
func (sheet *sheet) lineUse(lineI int) {
	sheet.lineMap[lineI] = struct{}{}
}

// isCoverLine 是否行数据已经被填充过
func (sheet *sheet) isCoverLine(lineI int) bool {
	_, has := sheet.lineMap[lineI]
	return has
}

// ExportInit 导出前的初始化
func (sheet *sheet) exportInit() error {
	if sheet.sheetName == "" {
		sheet.SetSheetName("Sheet1")
	}
	if sheet.DataBuilder == nil {
		sheet.SetDataBuilder(defaultDataBuilder)
	}
	// 获取数据
	sheet.heads = sheet.DataBuilder.GetHeads()
	var err error
	sheet.lines, err = sheet.DataBuilder.GetLines()
	if err != nil {
		return err
	}
	// 初始化file
	if sheet.file == nil {
		sheet.SetFile(NewFile().SetExcel(excelize.NewFile()))
	}
	sheet.lineMap = make(map[int]struct{})
	// 保持数字不转为科学技术法
	if sheet.numKeep {
		sheet.keepNum()
	}

	return nil
}

// keepNum 保持为数字（把数字转为string类型，输出的时候自然就是保持原数据了）
func (sheet *sheet) keepNum() {
	for i := range sheet.lines {
		for lineIndex := range sheet.lines[i] {
			is, s := util.NumToString(sheet.lines[i][lineIndex])
			if is {
				sheet.lines[i][lineIndex] = s
			}
		}
	}
}

// Export 导出
func (sheet *sheet) Export() error {
	// 先进行初始化
	err := sheet.exportInit()
	if err != nil {
		return err
	}
	f := sheet.file.f

	sheetIndex := f.NewSheet(sheet.sheetName) // 创建工作簿
	f.SetActiveSheet(sheetIndex)              // 设置激活的工作簿

	// 写入头信息
	if len(sheet.heads) > 0 {
		for headI, head := range sheet.heads {
			// 先填充头
			for index, headData := range head {
				lineIStr := strconv.FormatInt(int64(headI+1), 10)
				err := f.SetCellValue(sheet.sheetName, util.ToLine(index+1)+lineIStr, headData)
				if err != nil {
					return err
				}
				sheet.incMaxChar(index + 1)
			}
			sheet.incMaxLine(headI)
			sheet.lineUse(headI)
		}
	}

	// 写入每一行
	for i, line := range sheet.lines {
		// 检测行数，防止覆盖
		if sheet.isCoverLine(i) {
			return error2.NewLineCoverError(i)
		}
		for lineIndex := range line {
			lineIStr := strconv.FormatInt(int64(i), 10)
			err := f.SetCellValue(sheet.sheetName, util.ToLine(lineIndex+1)+lineIStr, line[lineIndex])
			if err != nil {
				return err
			}
			sheet.incMaxChar(lineIndex + 1)
		}
		sheet.incMaxLine(i)
		sheet.lineUse(i)
	}

	// 进行格式修饰
	return sheet.upStyle()
}

// upStyle 进行格式修饰
func (sheet *sheet) upStyle() error {
	// 格式化头
	if err := sheet.upHeadStyle(); err != nil {
		return err
	}
	// 格式化身体
	if err := sheet.upBodyStyle(); err != nil {
		return err
	}
	// 修饰宽度
	if err := sheet.upColWidth(); err != nil {
		return err
	}
	// 修饰高度
	if err := sheet.upColHeight(); err != nil {
		return err
	}
	// 修饰合并
	if err := sheet.upMergeCell(); err != nil {
		return err
	}
	return nil
}

// setStyle 样式设置
func (sheet *sheet) setStyle(style *Style) error {
	return setStyle(sheet.file.f, sheet.sheetName, style)
}

// upHeadStyle 修饰头信息
func (sheet *sheet) upHeadStyle() error {
	if len(sheet.heads) == 0 {
		return nil
	}
	headStyle := sheet.file.headStyle
	if sheet.headStyle != nil {
		headStyle = sheet.headStyle
	}
	style := NewStyleByNum(1, 1, sheet.maxChar, len(sheet.heads), headStyle)
	return sheet.setStyle(style)
}

// upBodyStyle 修饰行信息
func (sheet *sheet) upBodyStyle() error {
	for _, style := range sheet.bodyStyle {
		err := sheet.setStyle(style)
		if err != nil {
			return err
		}
	}
	return nil
}

// upColWidth 修饰宽度
func (sheet *sheet) upColWidth() error {
	for _, colWidth := range sheet.colWidth {
		err := setColWidth(sheet.file.f, sheet.sheetName, colWidth)
		if err != nil {
			return err
		}
	}
	return nil
}

// upColHeight 修饰高度
func (sheet *sheet) upColHeight() error {
	for _, colHeight := range sheet.colHeight {
		err := setColHeight(sheet.file.f, sheet.sheetName, colHeight)
		if err != nil {
			return err
		}
	}
	return nil
}

// upMergeCell 修饰合并
func (sheet *sheet) upMergeCell() error {
	for _, merge := range sheet.mergeCell {
		err := mergeCell(sheet.file.f, sheet.sheetName, merge)
		if err != nil {
			return err
		}
	}
	return nil
}

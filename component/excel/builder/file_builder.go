package builder

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/actorbuf/iota/generator/uuid"
	"github.com/gin-gonic/gin"
)

type File struct {
	f         *excelize.File
	fileName  string
	sheets    []*sheet
	headStyle *excelize.Style
	fileStyle []*Style
}

func NewFile() *File {
	return new(File)
}

// SetFileName 设置文件名（导出到web的时候才有用）
func (file *File) SetFileName(fileName string) *File {
	file.fileName = fileName
	return file
}

// AddSheet 添加sheet
func (file *File) AddSheet(sheet *sheet) *File {
	file.sheets = append(file.sheets, sheet)
	return file
}

// AddFileStyle 添加整体样式
func (file *File) AddFileStyle(style *Style) *File {
	if file.fileStyle == nil {
		file.fileStyle = make([]*Style, 0)
	}

	file.fileStyle = append(file.fileStyle, style)
	return file
}

// SetHeadStyle 设置头部样式，优先级比sheet低；sheet有设置则头部不生效
func (file *File) SetHeadStyle(style *excelize.Style) *File {
	file.headStyle = style
	return file
}

// SetExcel 植入excel文件，如果没有植入，会自动创建一个新的
func (file *File) SetExcel(f *excelize.File) *File {
	file.f = f
	return file
}

// exportInit 导出前初始化
func (file *File) exportInit() {
	if file.headStyle == nil {
		file.SetHeadStyle(&DefaultHeadStyle)
	}
	if file.f == nil {
		file.SetExcel(excelize.NewFile())
	}
}

// Export 导出
func (file *File) Export() (*excelize.File, error) {
	file.exportInit() // 统一初始化的地方
	for i := range file.sheets {
		file.sheets[i].SetFile(file) // 植入file
		err := file.sheets[i].Export()
		if err != nil {
			return nil, err
		}
	}
	return file.f, nil
}

// ExportFile 导出到文件
func (file *File) ExportFile(w io.Writer) error {
	f, err := file.Export()
	if err != nil {
		return err
	}
	return f.Write(w)
}

// ExportWeb 导出给web
func (file *File) ExportWeb(webContext ExportWebInterface) error {
	if file.fileName == "" {
		file.fileName = fmt.Sprintf("excel_%s.xlsx", time.Now().Format("20060102150405"))
	}
	if file.f == nil {
		_, err := file.Export()
		if err != nil {
			return err
		}
	}
	webContext.Header("content-description", "File Transfer")
	webContext.Header("content-type", "application/octet-stream")
	webContext.Header("content-disposition", "attachment; filename="+url.QueryEscape(filepath.Base(file.fileName)))
	webContext.Header("content-transfer-encoding", "binary")
	webContext.AddHeader("Access-Control-Expose-Headers", "content-disposition")
	webContext.Header("pragma", "public")
	webContext.Header("etag", uuid.TimeUUID().String())
	return file.f.Write(webContext.GetWriter())
}

// ExportGin 导出到Gin
func (file *File) ExportGin(ctx *gin.Context) error {
	web := ExportGin{ctx}
	return file.ExportWeb(&web)
}

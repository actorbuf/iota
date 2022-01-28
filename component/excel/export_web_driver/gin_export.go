package export_web_driver

import (
	"github.com/actorbuf/iota/component/excel/builder"
	"github.com/gin-gonic/gin"
	"io"
)

// ExportGin 输出到gin
type ExportGin struct {
	*gin.Context
	File *builder.File
}

// GetWriter 获取gin的writer
func (gin *ExportGin) GetWriter() io.Writer {
	return gin.Writer
}

// AddHeader 加入导出的特定头信息
func (gin *ExportGin) AddHeader(key, value string) {
	nameValues := gin.Writer.Header().Values(key)
	// 如果是"*"设置为当前值，如果不是，则增加
	if len(nameValues) == 1 && nameValues[0] == "*" {
		gin.Writer.Header().Set(key, value)
	} else {
		gin.Writer.Header().Add(key, value)
	}
}

// Export 导出到gin
func (gin *ExportGin) Export() error {
	return gin.File.ExportWeb(gin)
}

// NewExportGin 创建gin导出器
func NewExportGin(file *builder.File, ctx *gin.Context) *ExportGin {
	return &ExportGin{Context: ctx, File: file}
}

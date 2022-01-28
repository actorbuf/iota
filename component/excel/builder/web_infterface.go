package builder

import (
	"io"

	"github.com/gin-gonic/gin"
)

// ExportWebInterface web输出统一接口
type ExportWebInterface interface {
	GetWriter() io.Writer        // 输出io流
	Header(key, value string)    // 输出的头部信息预设值
	AddHeader(key, value string) // 主加输出的头部信息
}

// ExportGin 输出到gin
type ExportGin struct {
	*gin.Context
}

func (gin *ExportGin) GetWriter() io.Writer {
	return gin.Writer
}

func (gin *ExportGin) AddHeader(key, value string) {
	nameValues := gin.Writer.Header().Values(key)
	// 如果是"*"设置为当前值，如果不是，则增加
	if len(nameValues) == 1 && nameValues[0] == "*" {
		gin.Writer.Header().Set(key, value)
	} else {
		gin.Writer.Header().Add(key, value)
	}
}

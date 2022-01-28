package builder

import (
	"io"
)

// ExportWebInterface web输出统一接口
type ExportWebInterface interface {
	GetWriter() io.Writer        // 输出io流
	Header(key, value string)    // 输出的头部信息预设值
	AddHeader(key, value string) // 主加输出的头部信息
}

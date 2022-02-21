package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
)

type Context struct {
	*gin.Context
}

// SetBinaryFile 设置请求为文件下载
func (c *Context) SetBinaryFile(filename string, data []byte) {
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Transfer-Encoding", "binary")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, _ = c.Writer.Write(data)
}

// NewMockContext 新建一个模拟 core.Context
func NewMockContext() *Context {
	r := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(r)
	return &Context{Context: g}
}

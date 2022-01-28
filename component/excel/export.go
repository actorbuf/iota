package excel

import (
	"path/filepath"
	"time"

	"github.com/actorbuf/iota/generator/uuid"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func SetExportHeader(c *gin.Context, f *excelize.File, fileName string) {
	if fileName == "" {
		fileName = "excel_" + time.Now().Format("20060102150405")
	}
	c.Header("content-description", "File Transfer")
	c.Header("content-type", "application/octet-stream")
	c.Header("content-disposition", "attachment; filename="+filepath.Base(fileName))
	c.Header("content-transfer-encoding", "binary")
	c.Writer.Header().Add("Access-Control-Expose-Headers", "content-disposition")
	c.Header("pragma", "public")
	c.Header("etag", uuid.TimeUUID().String())
	_ = f.Write(c.Writer)
}

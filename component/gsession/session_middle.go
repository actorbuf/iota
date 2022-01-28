package gsession

import (
	"github.com/actorbuf/iota/component/gsession/driver"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SessionMiddleware session中间件
func SessionMiddleware(drive driver.Driver, attribute Attribute) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 前置操作操作，初始化session
		err := StartSession(c, drive, attribute)
		if err != nil {
			logrus.Errorf("get err: %v", err)
		}

		c.Next()
		// 后置操作，更新session
		session := GetSession(c)
		if session == nil {
			return
		}
		err = session.Save(c)
		if err != nil {
			logrus.Errorf("err: %v", err)
		}
	}
}

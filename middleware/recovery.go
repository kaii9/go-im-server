package middleware

import (
	"go-im-server/common"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Logger, _ = zap.NewProduction()

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error("panic recovered", zap.Any("error", err))
				common.Error(c, common.ErrInternal, common.GetErrMsg(common.ErrInternal))
				c.Abort()
			}
		}()
		c.Next()
	}
}

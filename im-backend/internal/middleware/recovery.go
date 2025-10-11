package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Recovery 恢复中间件 - 捕获panic并返回500错误
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic堆栈
				logrus.WithFields(logrus.Fields{
					"error":      fmt.Sprintf("%v", err),
					"stack":      string(debug.Stack()),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Error("Panic recovered")

				// 返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "服务器内部错误，请稍后重试",
					"error":   "INTERNAL_SERVER_ERROR",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}

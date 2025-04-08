// Package gin provides a unified response middleware for Gin framework.
// It standardizes API responses with code, message, and data fields.
package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ResponseMiddleware 响应中间件
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果已经设置了响应，则不再处理
		if c.Writer.Written() {
			return
		}

		// 获取响应状态码
		status := c.Writer.Status()
		if status == http.StatusOK {
			Success(c, nil)
		} else {
			Error(c, status, http.StatusText(status))
		}
	}
}

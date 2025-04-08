// Package gin provides a unified response middleware for Gin framework.
// It standardizes API responses with code, message, and data fields.
package gin

import (
	"net/http"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ResponseCode 响应码
type ResponseCode int

const (
	// CodeSuccess 成功
	CodeSuccess ResponseCode = 0
	// CodeError 错误
	CodeError ResponseCode = 1
	// CodeUnauthorized 未授权
	CodeUnauthorized ResponseCode = 401
	// CodeForbidden 禁止访问
	CodeForbidden ResponseCode = 403
	// CodeNotFound 资源不存在
	CodeNotFound ResponseCode = 404
	// CodeServerError 服务器错误
	CodeServerError ResponseCode = 500
)

// Response 统一响应结构
type Response struct {
	Code      errors.ErrorCode `json:"code"`                 // 响应码
	Message   string           `json:"message"`              // 响应消息
	Data      interface{}      `json:"data,omitempty"`       // 响应数据
	Timestamp int64            `json:"timestamp"`            // 时间戳
	RequestID string           `json:"request_id,omitempty"` // 请求ID
}

// Config 中间件配置
type Config struct {
	// 是否启用请求ID
	EnableRequestID bool
	// 请求ID的Header名称
	RequestIDHeader string
	// 自定义错误处理函数
	ErrorHandler func(*gin.Context, error)
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	EnableRequestID: true,
	RequestIDHeader: "X-Request-ID",
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response(c, errors.CodeSuccess, data)
}

// Error 错误响应
func Error(c *gin.Context, code errors.ErrorCode, message string) {
	response(c, code, nil, message)
}

// response 统一响应处理
func response(c *gin.Context, code errors.ErrorCode, data interface{}, messages ...string) {
	message := errors.GetMessage(code)
	if len(messages) > 0 {
		message = messages[0]
	}

	resp := &Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if DefaultConfig.EnableRequestID {
		if requestID := c.GetHeader(DefaultConfig.RequestIDHeader); requestID != "" {
			resp.RequestID = requestID
		}
	}

	c.JSON(http.StatusOK, resp)
}

// ResponseMiddleware 响应中间件
func ResponseMiddleware(config ...*Config) gin.HandlerFunc {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 设置请求ID
		if cfg.EnableRequestID {
			requestID := c.GetHeader(cfg.RequestIDHeader)
			if requestID == "" {
				requestID = generateRequestID()
				c.Header(cfg.RequestIDHeader, requestID)
			}
		}

		// 错误处理
		c.Next()

		// 如果已经设置了响应，则不再处理
		if c.Writer.Written() {
			return
		}

		// 获取最后一个错误
		lastError := c.Errors.Last()
		if lastError != nil {
			if cfg.ErrorHandler != nil {
				cfg.ErrorHandler(c, lastError)
			} else {
				if e, ok := lastError.Err.(*errors.Error); ok {
					Error(c, e.Code, e.Message)
				} else {
					Error(c, errors.CodeServerError, lastError.Error())
				}
			}
			return
		}

		// 获取响应状态码
		status := c.Writer.Status()
		switch status {
		case http.StatusOK:
			Success(c, nil)
		case http.StatusUnauthorized:
			Error(c, errors.CodeUnauthorized, "")
		case http.StatusForbidden:
			Error(c, errors.CodeForbidden, "")
		case http.StatusNotFound:
			Error(c, errors.CodeNotFound, "")
		default:
			Error(c, errors.CodeServerError, http.StatusText(status))
		}
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

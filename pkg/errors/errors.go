package errors

import (
	"fmt"
	"strings"
)

// ErrorCode 错误码
type ErrorCode int

const (
	// CodeSuccess 成功
	CodeSuccess ErrorCode = 0
	// CodeError 错误
	CodeError ErrorCode = 1
	// CodeUnauthorized 未授权
	CodeUnauthorized ErrorCode = 401
	// CodeForbidden 禁止访问
	CodeForbidden ErrorCode = 403
	// CodeNotFound 资源不存在
	CodeNotFound ErrorCode = 404
	// CodeServerError 服务器错误
	CodeServerError ErrorCode = 500
)

// Error 自定义错误类型
type Error struct {
	Code    ErrorCode `json:"code"`    // 错误码
	Message string    `json:"message"` // 错误消息
	Details []string  `json:"details"` // 错误详情
}

// New 创建新的错误
func New(code ErrorCode, message string, details ...string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s", e.Message, strings.Join(e.Details, "; "))
	}
	return e.Message
}

// WithDetails 添加错误详情
func (e *Error) WithDetails(details ...string) *Error {
	e.Details = append(e.Details, details...)
	return e
}

// Is 判断错误是否匹配
func (e *Error) Is(target error) bool {
	if target == nil {
		return false
	}
	if err, ok := target.(*Error); ok {
		return e.Code == err.Code
	}
	return false
}

// DefaultErrorMessages 默认错误消息映射
var DefaultErrorMessages = map[ErrorCode]string{
	CodeSuccess:      "success",
	CodeError:        "error",
	CodeUnauthorized: "unauthorized",
	CodeForbidden:    "forbidden",
	CodeNotFound:     "not found",
	CodeServerError:  "server error",
}

// GetMessage 获取错误消息
func GetMessage(code ErrorCode) string {
	if msg, ok := DefaultErrorMessages[code]; ok {
		return msg
	}
	return "unknown error"
}

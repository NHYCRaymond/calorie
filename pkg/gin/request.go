// Package gin provides a unified request middleware for Gin framework.
// It handles common request processing tasks such as logging, tracing,
// rate limiting, and security headers.
package gin

import (
	"context"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// 常量定义
const (
	headerContentTypeOptions = "X-Content-Type-Options"
	headerFrameOptions       = "X-Frame-Options"
	headerXSSProtection      = "X-XSS-Protection"
	headerHSTS               = "Strict-Transport-Security"
	defaultMaxBodySize       = 10 << 20 // 10MB
)

// 对象池
var (
	requestFieldsPool = sync.Pool{
		New: func() interface{} {
			return make(logrus.Fields, 10)
		},
	}
	requestCounter uint64
	logChan        = make(chan logrus.Fields, 1000)
)

// RequestConfig 请求配置
type RequestConfig struct {
	// 服务名称
	ServiceName string
	// 请求超时时间
	Timeout time.Duration
	// 请求ID的Header名称
	RequestIDHeader string
	// 是否启用请求日志
	EnableRequestLog bool
	// 是否启用安全头
	EnableSecurityHeaders bool
	// 请求限流配置
	RateLimit int
	RateBurst int
	// 请求大小限制
	MaxBodySize int64
	// 路径过滤
	PathWhitelist []string
	PathBlacklist []string
	// 请求头过滤
	FilterHeaders []string
	// 自定义标签
	Labels map[string]string
	// 自定义日志字段
	LogFields logrus.Fields
}

// DefaultRequestConfig 默认配置
var DefaultRequestConfig = &RequestConfig{
	ServiceName:           "default",
	Timeout:               30 * time.Second,
	RequestIDHeader:       "X-Request-ID",
	EnableRequestLog:      true,
	EnableSecurityHeaders: true,
	RateLimit:             1000,
	RateBurst:             100,
	MaxBodySize:           defaultMaxBodySize,
	Labels:                make(map[string]string),
	LogFields:             make(logrus.Fields),
}

// RequestMiddleware 请求处理中间件
func RequestMiddleware(config *RequestConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRequestConfig
	}

	// 初始化限流器
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), config.RateBurst)

	// 启动日志处理协程
	go processLogs()

	return func(c *gin.Context) {
		// 1. 请求大小限制
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxBodySize)

		// 2. 路径过滤
		if !isPathAllowed(c.Request.URL.Path, config) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Access Denied",
			})
			return
		}

		// 3. 限流检查
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "Too Many Requests",
			})
			return
		}

		// 4. 添加请求上下文和超时控制
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		// 5. 添加请求ID
		requestID := c.GetHeader(config.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header(config.RequestIDHeader, requestID)
		}

		// 6. 添加安全头
		if config.EnableSecurityHeaders {
			c.Header(headerContentTypeOptions, "nosniff")
			c.Header(headerFrameOptions, "DENY")
			c.Header(headerXSSProtection, "1; mode=block")
			c.Header(headerHSTS, "max-age=31536000; includeSubDomains")
		}

		// 7. Panic 恢复
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logFields := requestFieldsPool.Get().(logrus.Fields)
				logFields["service"] = config.ServiceName
				logFields["request_id"] = requestID
				logFields["error"] = err
				logFields["stack"] = string(stack)
				logChan <- logFields
				requestFieldsPool.Put(logFields)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "Internal Server Error",
				})
			}
		}()

		// 8. 记录开始时间
		start := time.Now()

		// 9. 处理请求
		c.Next()

		// 10. 记录请求日志
		if config.EnableRequestLog {
			duration := time.Since(start)
			path := c.Request.URL.Path
			raw := c.Request.URL.RawQuery
			if raw != "" {
				path = path + "?" + raw
			}

			// 使用对象池获取字段
			fields := requestFieldsPool.Get().(logrus.Fields)
			fields["service"] = config.ServiceName
			fields["request_id"] = requestID
			fields["client_ip"] = c.ClientIP()
			fields["method"] = c.Request.Method
			fields["path"] = path
			fields["status"] = c.Writer.Status()
			fields["latency_ms"] = float64(duration.Nanoseconds()) / 1e6
			fields["user_agent"] = c.Request.UserAgent()
			fields["error_count"] = len(c.Errors)
			fields["request_count"] = atomic.AddUint64(&requestCounter, 1)

			// 添加自定义字段
			for k, v := range config.LogFields {
				fields[k] = v
			}

			// 异步处理日志
			logChan <- fields
		}

		// 11. 处理上下文超时
		if ctx.Err() == context.DeadlineExceeded {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"code":    http.StatusGatewayTimeout,
				"message": "Request Timeout",
			})
			return
		}
	}
}

// isPathAllowed 检查路径是否允许访问
func isPathAllowed(path string, config *RequestConfig) bool {
	// 检查黑名单
	for _, blackPath := range config.PathBlacklist {
		if path == blackPath {
			return false
		}
	}

	// 如果白名单不为空，只允许白名单中的路径
	if len(config.PathWhitelist) > 0 {
		for _, whitePath := range config.PathWhitelist {
			if path == whitePath {
				return true
			}
		}
		return false
	}

	return true
}

// processLogs 处理日志的协程
func processLogs() {
	for fields := range logChan {
		// 根据状态码选择日志级别
		status, _ := fields["status"].(int)
		logger := logrus.WithFields(fields)
		if status >= http.StatusInternalServerError {
			logger.Error("请求处理失败")
		} else if status >= http.StatusBadRequest {
			logger.Warn("请求处理异常")
		} else {
			logger.Info("请求处理完成")
		}
		// 归还对象到池中
		requestFieldsPool.Put(fields)
	}
}

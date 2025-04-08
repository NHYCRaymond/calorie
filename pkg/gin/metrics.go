package gin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NHYCRaymond/calorie/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsConfig Prometheus 指标配置
type MetricsConfig struct {
	// 是否启用指标收集
	Enabled bool
	// 服务名称
	ServiceName string
	// 自定义标签
	Labels map[string]string
	// 是否记录请求体大小
	EnableRequestSize bool
	// 是否记录响应体大小
	EnableResponseSize bool
	// 是否记录请求成功率
	EnableSuccessRate bool
	// 是否记录请求失败率
	EnableFailureRate bool
	// 是否记录错误类型
	EnableErrorType bool
	// 是否记录并发请求数
	EnableConcurrentRequests bool
	// 是否记录请求延迟分布
	EnableLatencyDistribution bool
	// 是否记录速率限制
	EnableRateLimit bool
	// 是否记录自定义业务指标
	EnableCustomMetrics bool
	// 是否记录请求上下文信息
	EnableRequestContext bool
	// 自定义指标收集器
	CustomCollector CustomMetricsCollector
}

// CustomMetricsCollector 自定义指标收集器接口
type CustomMetricsCollector interface {
	Collect(c *gin.Context, statusCode int, duration time.Duration)
}

// DefaultMetricsConfig 默认配置
var DefaultMetricsConfig = &MetricsConfig{
	Enabled:                   true,
	ServiceName:               "default",
	Labels:                    make(map[string]string),
	EnableRequestSize:         true,
	EnableResponseSize:        true,
	EnableSuccessRate:         true,
	EnableFailureRate:         true,
	EnableErrorType:           true,
	EnableConcurrentRequests:  true,
	EnableLatencyDistribution: true,
	EnableRateLimit:           true,
	EnableCustomMetrics:       false,
	EnableRequestContext:      false,
}

// PrometheusMiddleware Prometheus 指标中间件
func PrometheusMiddleware(client *metrics.Client, config *MetricsConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultMetricsConfig
	}

	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 创建请求计数器
	requestCounter := client.Counter(
		"http_requests_total",
		"Total number of HTTP requests",
		[]string{"method", "path", "status", "service"},
	)

	// 创建请求持续时间直方图
	requestDuration := client.Histogram(
		"http_request_duration_seconds",
		"HTTP request duration in seconds",
		[]string{"method", "path", "service"},
		prometheus.DefBuckets,
	)

	// 创建并发请求数指标
	var concurrentRequests *prometheus.GaugeVec
	if config.EnableConcurrentRequests {
		concurrentRequests = client.Gauge(
			"http_concurrent_requests",
			"Number of concurrent HTTP requests",
			[]string{"method", "path", "service"},
		)
	}

	// 创建请求延迟分布指标
	var latencyDistribution *prometheus.HistogramVec
	if config.EnableLatencyDistribution {
		latencyDistribution = client.Histogram(
			"http_request_latency_seconds",
			"HTTP request latency distribution",
			[]string{"method", "path", "service"},
			[]float64{0.1, 0.5, 1, 2, 5, 10},
		)
	}

	// 创建请求体大小直方图
	var requestSize *prometheus.HistogramVec
	if config.EnableRequestSize {
		requestSize = client.Histogram(
			"http_request_size_bytes",
			"HTTP request size in bytes",
			[]string{"method", "path", "service"},
			prometheus.ExponentialBuckets(100, 10, 8),
		)
	}

	// 创建响应体大小直方图
	var responseSize *prometheus.HistogramVec
	if config.EnableResponseSize {
		responseSize = client.Histogram(
			"http_response_size_bytes",
			"HTTP response size in bytes",
			[]string{"method", "path", "status", "service"},
			prometheus.ExponentialBuckets(100, 10, 8),
		)
	}

	// 创建请求成功率计数器
	var successCounter *prometheus.CounterVec
	if config.EnableSuccessRate {
		successCounter = client.Counter(
			"http_requests_success_total",
			"Total number of successful HTTP requests",
			[]string{"method", "path", "service"},
		)
	}

	// 创建请求失败率计数器
	var failureCounter *prometheus.CounterVec
	if config.EnableFailureRate {
		failureCounter = client.Counter(
			"http_requests_failure_total",
			"Total number of failed HTTP requests",
			[]string{"method", "path", "status", "service"},
		)
	}

	// 创建错误类型计数器
	var errorTypeCounter *prometheus.CounterVec
	if config.EnableErrorType {
		errorTypeCounter = client.Counter(
			"http_requests_error_type_total",
			"Total number of errors by type",
			[]string{"method", "path", "error_type", "service"},
		)
	}

	// 创建速率限制指标
	var rateLimitCounter *prometheus.CounterVec
	if config.EnableRateLimit {
		rateLimitCounter = client.Counter(
			"http_rate_limit_total",
			"Total number of rate limited requests",
			[]string{"method", "path", "service"},
		)
	}

	// 创建请求上下文指标
	var requestContextCounter *prometheus.CounterVec
	if config.EnableRequestContext {
		requestContextCounter = client.Counter(
			"http_request_context_total",
			"Total number of requests with context information",
			[]string{"method", "path", "service", "context_key"},
		)
	}

	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 增加并发请求数
		if config.EnableConcurrentRequests && concurrentRequests != nil {
			concurrentRequests.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				config.ServiceName,
			).Inc()
			defer concurrentRequests.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				config.ServiceName,
			).Dec()
		}

		// 记录请求体大小
		if config.EnableRequestSize && requestSize != nil {
			requestSize.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				config.ServiceName,
			).Observe(float64(c.Request.ContentLength))
		}

		// 处理请求
		c.Next()

		// 计算持续时间
		duration := time.Since(start).Seconds()

		// 获取状态码
		status := fmt.Sprintf("%d", c.Writer.Status())
		statusCode := c.Writer.Status()

		// 记录请求计数
		requestCounter.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			status,
			config.ServiceName,
		).Inc()

		// 记录请求持续时间
		requestDuration.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			config.ServiceName,
		).Observe(duration)

		// 记录响应体大小
		if config.EnableResponseSize && responseSize != nil {
			responseSize.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				status,
				config.ServiceName,
			).Observe(float64(c.Writer.Size()))
		}

		// 记录成功请求
		if config.EnableSuccessRate && successCounter != nil && statusCode < http.StatusBadRequest {
			successCounter.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				config.ServiceName,
			).Inc()
		}

		// 记录失败请求
		if config.EnableFailureRate && failureCounter != nil && statusCode >= http.StatusBadRequest {
			failureCounter.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				status,
				config.ServiceName,
			).Inc()
		}

		// 记录错误类型
		if config.EnableErrorType && errorTypeCounter != nil && statusCode >= http.StatusBadRequest {
			errorType := getErrorType(statusCode)
			errorTypeCounter.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				errorType,
				config.ServiceName,
			).Inc()
		}

		// 记录请求延迟分布
		if config.EnableLatencyDistribution && latencyDistribution != nil {
			latencyDistribution.WithLabelValues(
				c.Request.Method,
				c.Request.URL.Path,
				config.ServiceName,
			).Observe(duration)
		}

		// 记录速率限制
		if config.EnableRateLimit && rateLimitCounter != nil {
			if isRateLimited, exists := c.Get("rate_limited"); exists {
				if isLimited, ok := isRateLimited.(bool); ok && isLimited {
					rateLimitCounter.WithLabelValues(
						c.Request.Method,
						c.Request.URL.Path,
						config.ServiceName,
					).Inc()
				}
			}
		}

		// 记录请求上下文信息
		if config.EnableRequestContext && requestContextCounter != nil {
			// 从上下文中获取自定义信息
			if contextKeys, exists := c.Get("context_keys"); exists {
				if keys, ok := contextKeys.([]string); ok {
					for _, key := range keys {
						if _, exists := c.Get(key); exists {
							requestContextCounter.WithLabelValues(
								c.Request.Method,
								c.Request.URL.Path,
								config.ServiceName,
								key,
							).Inc()
						}
					}
				}
			}
		}

		// 执行自定义指标收集
		if config.EnableCustomMetrics && config.CustomCollector != nil {
			config.CustomCollector.Collect(c, statusCode, time.Duration(duration*float64(time.Second)))
		}
	}
}

// getErrorType 根据状态码获取错误类型
func getErrorType(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode == 429:
		return "rate_limit"
	case statusCode == 401:
		return "unauthorized"
	case statusCode == 403:
		return "forbidden"
	case statusCode == 404:
		return "not_found"
	case statusCode == 400:
		return "bad_request"
	default:
		return "other_error"
	}
}

// ExampleCustomCollector 示例自定义指标收集器
type ExampleCustomCollector struct {
	client *metrics.Client
}

// NewExampleCustomCollector 创建示例自定义指标收集器
func NewExampleCustomCollector(client *metrics.Client) *ExampleCustomCollector {
	return &ExampleCustomCollector{
		client: client,
	}
}

// Collect 实现自定义指标收集
func (c *ExampleCustomCollector) Collect(ctx *gin.Context, statusCode int, duration time.Duration) {
	// 示例：记录业务特定指标
	customCounter := c.client.Counter(
		"custom_business_metric_total",
		"Custom business metric",
		[]string{"method", "path", "status", "service"},
	)

	customCounter.WithLabelValues(
		ctx.Request.Method,
		ctx.Request.URL.Path,
		fmt.Sprintf("%d", statusCode),
		"example_service",
	).Inc()
}

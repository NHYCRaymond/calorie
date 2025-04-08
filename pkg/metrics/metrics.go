package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config 监控配置
type Config struct {
	// 是否启用监控
	Enabled bool
	// 监控服务地址
	Addr string
	// 监控路径
	Path string
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	Enabled: true,
	Addr:    ":9090",
	Path:    "/metrics",
}

// Client Prometheus 客户端
type Client struct {
	config *Config

	// 计数器
	counters map[string]*prometheus.CounterVec
	// 仪表盘
	gauges map[string]*prometheus.GaugeVec
	// 直方图
	histograms map[string]*prometheus.HistogramVec
	// 摘要
	summaries map[string]*prometheus.SummaryVec
}

// NewClient 创建新的监控客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig
	}

	client := &Client{
		config:     config,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}

	if config.Enabled {
		// 启动 Prometheus HTTP 服务
		go func() {
			http.Handle(config.Path, promhttp.Handler())
			http.ListenAndServe(config.Addr, nil)
		}()
	}

	return client
}

// Counter 创建或获取计数器
func (c *Client) Counter(name, help string, labels []string) *prometheus.CounterVec {
	if counter, ok := c.counters[name]; ok {
		return counter
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister(counter)
	c.counters[name] = counter
	return counter
}

// Gauge 创建或获取仪表盘
func (c *Client) Gauge(name, help string, labels []string) *prometheus.GaugeVec {
	if gauge, ok := c.gauges[name]; ok {
		return gauge
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister(gauge)
	c.gauges[name] = gauge
	return gauge
}

// Histogram 创建或获取直方图
func (c *Client) Histogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	if histogram, ok := c.histograms[name]; ok {
		return histogram
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	prometheus.MustRegister(histogram)
	c.histograms[name] = histogram
	return histogram
}

// Summary 创建或获取摘要
func (c *Client) Summary(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	if summary, ok := c.summaries[name]; ok {
		return summary
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)
	prometheus.MustRegister(summary)
	c.summaries[name] = summary
	return summary
}

// 预定义一些常用的指标
var (
	// HTTP 请求计数
	HTTPRequestTotal = "http_requests_total"
	// HTTP 请求延迟
	HTTPRequestDuration = "http_request_duration_seconds"
	// 内存使用
	MemoryUsage = "memory_usage_bytes"
	// CPU 使用
	CPUUsage = "cpu_usage_percent"
)

// 预定义一些常用的标签
var (
	// HTTP 方法标签
	LabelMethod = "method"
	// HTTP 路径标签
	LabelPath = "path"
	// HTTP 状态码标签
	LabelStatus = "status"
	// 服务名称标签
	LabelService = "service"
)

// Package logger provides a logrus-based logging utility with rotation support.
// It supports daily rotation and size-based rotation (default 64MB).
package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log *logrus.Logger
	mu  sync.RWMutex
)

// Config 日志配置
type Config struct {
	LogPath     string // 日志文件路径
	MaxSize     int    // 单个日志文件最大大小（MB）
	MaxBackups  int    // 保留的旧日志文件最大数量
	MaxAge      int    // 保留的旧日志文件最大天数
	Compress    bool   // 是否压缩旧日志文件
	Level       string // 日志级别
	ServiceName string // 服务名称
}

// InitLogger 初始化日志配置
func InitLogger(config *Config) {
	mu.Lock()
	defer mu.Unlock()

	if log != nil {
		return
	}

	log = logrus.New()

	// 设置日志格式为 JSON
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// 确保日志目录存在
	if err := os.MkdirAll(config.LogPath, 0755); err != nil {
		logrus.Fatal("Failed to create log directory:", err)
	}

	// 创建日志文件
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(config.LogPath, "app.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// 同时输出到文件和控制台
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// 添加服务名称字段
	if config.ServiceName != "" {
		log = log.WithField("service", config.ServiceName).Logger
	}
}

// Debug 输出调试日志
func Debug(args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		log.Debug(args...)
	}
}

// Info 输出信息日志
func Info(args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		log.Info(args...)
	}
}

// Warn 输出警告日志
func Warn(args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		log.Warn(args...)
	}
}

// Error 输出错误日志
func Error(args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		log.Error(args...)
	}
}

// Fatal 输出致命错误日志
func Fatal(args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		log.Fatal(args...)
	}
}

// WithFields 添加字段到日志
func WithFields(fields logrus.Fields) *logrus.Entry {
	mu.RLock()
	defer mu.RUnlock()
	if log != nil {
		return log.WithFields(fields)
	}
	return logrus.NewEntry(logrus.New())
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	LogPath:     "logs",
	MaxSize:     64,
	MaxBackups:  10,
	MaxAge:      30,
	Compress:    true,
	Level:       "info",
	ServiceName: "default",
}

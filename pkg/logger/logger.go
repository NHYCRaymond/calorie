// Package logger provides a logrus-based logging utility with rotation support.
// It supports daily rotation and size-based rotation (default 64MB).
package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

// Config 日志配置
type Config struct {
	LogPath    string // 日志文件路径
	MaxSize    int    // 单个日志文件最大大小（MB）
	MaxBackups int    // 保留的旧日志文件最大数量
	MaxAge     int    // 保留的旧日志文件最大天数
	Compress   bool   // 是否压缩旧日志文件
}

// InitLogger 初始化日志配置
func InitLogger(config *Config) {
	if log != nil {
		return
	}

	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// 确保日志目录存在
	if err := os.MkdirAll(config.LogPath, 0755); err != nil {
		panic(err)
	}

	// 设置日志输出
	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join(config.LogPath, "app.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	})
}

// Debug 输出调试日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info 输出信息日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Warn 输出警告日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error 输出错误日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Fatal 输出致命错误日志
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// WithFields 添加字段到日志
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

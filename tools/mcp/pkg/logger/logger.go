package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/internal/config"
)

// Logger 全局日志实例
var Logger *logrus.Logger

// Init 初始化日志系统
func Init(cfg *config.LoggingConfig) error {
	Logger = logrus.New()
	
	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	Logger.SetLevel(level)
	
	// 设置日志格式
	switch cfg.Format {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	
	// 设置输出目标
	if cfg.File != "" {
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
			return err
		}
		
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		
		// 同时输出到文件和控制台
		Logger.SetOutput(io.MultiWriter(os.Stdout, file))
	} else {
		Logger.SetOutput(os.Stdout)
	}
	
	return nil
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if Logger == nil {
		// 如果没有初始化，使用默认配置
		Logger = logrus.New()
		Logger.SetLevel(logrus.InfoLevel)
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	return Logger
}

// WithField 创建带字段的日志条目
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields 创建带多个字段的日志条目
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError 创建带错误信息的日志条目
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// Debug 记录调试级别日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 记录格式化的调试级别日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 记录信息级别日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 记录格式化的信息级别日志
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 记录警告级别日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 记录格式化的警告级别日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 记录错误级别日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 记录格式化的错误级别日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 记录致命错误日志并退出程序
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 记录格式化的致命错误日志并退出程序
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic 记录恐慌日志并触发panic
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf 记录格式化的恐慌日志并触发panic
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}
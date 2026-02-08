package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger
var requestLogger *zap.Logger

// InitLogger 初始化日志
func InitLogger(level, logPath string) error {
	// 解析日志级别
	zapLevel := parseLevel(level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	cores := []zapcore.Core{}

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)
	cores = append(cores, consoleCore)

	// 文件输出
	if logPath != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 配置日志轮转
		fileWriter := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    100, // MB
			MaxBackups: 10,
			MaxAge:     30, // days
			Compress:   true,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(fileWriter),
			zapLevel,
		)
		cores = append(cores, fileCore)
	}

	// 创建logger
	core := zapcore.NewTee(cores...)
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetLogger 获取全局logger
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置
		globalLogger, _ = zap.NewProduction()
	}
	return globalLogger
}

// Sync 同步日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With 创建带字段的logger
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// InitRequestLogger 初始化请求日志记录器（按日期分文件）
func InitRequestLogger(logDir string) error {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 生成当天的日志文件名
	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", today))

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置日志轮转
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // MB
		MaxBackups: 30,  // 保留30个备份
		MaxAge:     30,  // 保留30天
		Compress:   true,
	}

	// 创建文件核心
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileCore := zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(fileWriter),
		zapcore.InfoLevel,
	)

	// 创建请求日志记录器
	requestLogger = zap.New(fileCore, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

// GetRequestLogger 获取请求日志记录器
func GetRequestLogger() *zap.Logger {
	if requestLogger == nil {
		// 如果未初始化，使用全局logger
		return GetLogger()
	}
	return requestLogger
}

// LogRequest 记录HTTP请求日志到日期文件
func LogRequest(msg string, fields ...zap.Field) {
	GetRequestLogger().Info(msg, fields...)
}

// UpdateRequestLoggerFile 更新请求日志文件（用于日期变更时）
func UpdateRequestLoggerFile(logDir string) error {
	return InitRequestLogger(logDir)
}

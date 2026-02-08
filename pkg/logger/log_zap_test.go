package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInitRequestLogger(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "test_logs")
	defer os.RemoveAll(tempDir)

	// 初始化请求日志记录器
	err := InitRequestLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize request logger: %v", err)
	}

	// 记录一些测试日志
	LogRequest("Test request 1",
		zap.String("method", "GET"),
		zap.String("path", "/api/test"),
		zap.Int("status", 200),
	)

	LogRequest("Test request 2",
		zap.String("method", "POST"),
		zap.String("path", "/api/users"),
		zap.Int("status", 201),
	)

	// 验证日志文件是否创建
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(tempDir, today+".log")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", logFile)
	}
}

func TestLogRequest(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "test_request_logs")
	defer os.RemoveAll(tempDir)

	// 初始化
	err := InitRequestLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize request logger: %v", err)
	}

	// 测试记录日志
	LogRequest("HTTP Request",
		zap.String("method", "GET"),
		zap.String("path", "/api/products"),
		zap.String("query", "page=1&size=10"),
		zap.Int("status", 200),
		zap.String("ip", "127.0.0.1"),
		zap.Duration("cost", 50*time.Millisecond),
	)

	// 验证文件存在
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(tempDir, today+".log")

	info, err := os.Stat(logFile)
	if err != nil {
		t.Errorf("Failed to stat log file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Log file is empty")
	}
}

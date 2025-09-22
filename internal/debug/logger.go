package debug

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger 调试日志器
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var (
	// 全局日志器
	GlobalLogger *Logger
)

// 初始化全局日志器
func init() {
	GlobalLogger = NewLogger(INFO)
}

// NewLogger 创建新的日志器
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetLevelFromString 从字符串设置日志级别
func (l *Logger) SetLevelFromString(levelStr string) {
	switch strings.ToLower(levelStr) {
	case "debug":
		l.SetLevel(DEBUG)
	case "info":
		l.SetLevel(INFO)
	case "warn", "warning":
		l.SetLevel(WARN)
	case "error":
		l.SetLevel(ERROR)
	default:
		l.SetLevel(INFO)
	}
}

// Debug 记录调试信息
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logWithLevel("DEBUG", format, args...)
	}
}

// Info 记录信息
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.logWithLevel("INFO", format, args...)
	}
}

// Warn 记录警告
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.logWithLevel("WARN", format, args...)
	}
}

// Error 记录错误
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.logWithLevel("ERROR", format, args...)
	}
}

// logWithLevel 带级别的日志记录
func (l *Logger) logWithLevel(level, format string, args ...interface{}) {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 简化文件路径
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	// 格式化消息
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05.000")

	// 输出日志
	fmt.Fprintf(os.Stderr, "[%s] %s %s:%d %s\n",
		timestamp, level, file, line, message)
}

// 全局函数，使用全局日志器

// Debug 全局调试日志
func Debug(format string, args ...interface{}) {
	GlobalLogger.Debug(format, args...)
}

// Info 全局信息日志
func Info(format string, args ...interface{}) {
	GlobalLogger.Info(format, args...)
}

// Warn 全局警告日志
func Warn(format string, args ...interface{}) {
	GlobalLogger.Warn(format, args...)
}

// Error 全局错误日志
func Error(format string, args ...interface{}) {
	GlobalLogger.Error(format, args...)
}

// SetLevel 设置全局日志级别
func SetLevel(level LogLevel) {
	GlobalLogger.SetLevel(level)
}

// SetLevelFromString 从字符串设置全局日志级别
func SetLevelFromString(levelStr string) {
	GlobalLogger.SetLevelFromString(levelStr)
}

// 从环境变量初始化日志级别
func init() {
	if level := os.Getenv("DEBUG_LEVEL"); level != "" {
		SetLevelFromString(level)
	} else if os.Getenv("DEBUG") == "true" {
		SetLevel(DEBUG)
	}
}

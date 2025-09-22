package debug

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	// 创建测试日志器
	logger := NewLogger(DEBUG)

	// 测试不同级别的日志
	logger.Debug("这是调试信息")
	logger.Info("这是信息")
	logger.Warn("这是警告")
	logger.Error("这是错误")

	// 测试全局日志器
	SetLevel(DEBUG)
	Debug("全局调试信息")
	Info("全局信息")
	Warn("全局警告")
	Error("全局错误")
}

func TestLoggerLevels(t *testing.T) {
	logger := NewLogger(INFO)

	// 测试级别过滤
	logger.Debug("这条调试信息不应该显示")
	logger.Info("这条信息应该显示")
	logger.Warn("这条警告应该显示")
	logger.Error("这条错误应该显示")
}

func TestLoggerFromString(t *testing.T) {
	logger := NewLogger(INFO)

	// 测试从字符串设置级别
	logger.SetLevelFromString("debug")
	if logger.level != DEBUG {
		t.Errorf("期望级别为DEBUG，实际为%d", logger.level)
	}

	logger.SetLevelFromString("error")
	if logger.level != ERROR {
		t.Errorf("期望级别为ERROR，实际为%d", logger.level)
	}

	logger.SetLevelFromString("invalid")
	if logger.level != INFO {
		t.Errorf("期望级别为INFO，实际为%d", logger.level)
	}
}

func TestGlobalLogger(t *testing.T) {
	// 测试全局日志器设置
	SetLevelFromString("debug")
	SetLevelFromString("warn")
	SetLevelFromString("error")
}

func TestEnvironmentVariables(t *testing.T) {
	// 测试环境变量
	os.Setenv("DEBUG", "true")
	os.Setenv("DEBUG_LEVEL", "warn")

	// 测试全局日志器设置
	SetLevelFromString("warn")

	// 清理环境变量
	os.Unsetenv("DEBUG")
	os.Unsetenv("DEBUG_LEVEL")
}

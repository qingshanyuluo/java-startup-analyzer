package main

import (
	"context"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/devops"
	"github.com/user/java-startup-analyzer/cmd"
	"github.com/user/java-startup-analyzer/internal/debug"
)

func main() {
	// 初始化调试功能
	initDebug()

	// 执行主程序
	cmd.Execute()
}

// initDebug 初始化调试功能
func initDebug() {
	// 检查是否启用调试模式
	if os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG_PROFILER") == "true" {
		// 设置调试日志级别
		if level := os.Getenv("DEBUG_LEVEL"); level != "" {
			debug.SetLevelFromString(level)
		} else {
			debug.SetLevel(debug.DEBUG)
		}

		debug.Info("调试模式已启用")

		// 启动性能分析器
		if os.Getenv("DEBUG_PROFILER") == "true" {
			port := os.Getenv("DEBUG_PROFILER_PORT")
			if port == "" {
				port = "6060"
			}

			go func() {
				if err := debug.StartProfiler(port); err != nil {
					debug.Error("启动性能分析器失败: %v", err)
				}
			}()

			debug.Info("性能分析器已启动: http://localhost:%s", port)
		}

		ctx := context.Background()

		// 1.调用调试服务初始化函数
		err := devops.Init(ctx)
		if err != nil {
			debug.Error("[eino dev] init failed, err=%v", err)
			return
		}

	}

	// 检查是否启用详细模式
	if os.Getenv("VERBOSE") == "true" || strings.Contains(strings.Join(os.Args, " "), "--verbose") {
		debug.SetLevel(debug.INFO)
		debug.Info("详细模式已启用")
	}
}

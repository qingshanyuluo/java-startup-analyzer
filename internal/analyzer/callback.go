package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// JavaAnalyzerCallback 用于记录Java分析器的详细执行过程
type JavaAnalyzerCallback struct {
	callbacks.HandlerBuilder
	logFile   *os.File
	logPath   string
	startTime time.Time
}

// NewJavaAnalyzerCallback 创建新的回调处理器
func NewJavaAnalyzerCallback(logDir string) (*JavaAnalyzerCallback, error) {
	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(logDir) {
		absLogDir, err := filepath.Abs(logDir)
		if err != nil {
			return nil, fmt.Errorf("无法解析日志目录路径: %w", err)
		}
		logDir = absLogDir
	}

	// 确保日志目录存在，包括所有父目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 创建带时间戳的日志文件
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("java_analyzer_%s.log", timestamp))

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %w", err)
	}

	return &JavaAnalyzerCallback{
		logFile: logFile,
		logPath: logPath,
	}, nil
}

// Close 关闭日志文件
func (cb *JavaAnalyzerCallback) Close() error {
	if cb.logFile != nil {
		return cb.logFile.Close()
	}
	return nil
}

// GetLogPath 获取日志文件路径
func (cb *JavaAnalyzerCallback) GetLogPath() string {
	return cb.logPath
}

// writeLog 写入日志的辅助方法
func (cb *JavaAnalyzerCallback) writeLog(level, message string, data interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	var logEntry strings.Builder
	logEntry.WriteString(fmt.Sprintf("[%s] [%s] %s", timestamp, level, message))

	if data != nil {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			logEntry.WriteString(fmt.Sprintf(" (序列化失败: %v)", err))
		} else {
			logEntry.WriteString("\n")
			logEntry.WriteString(string(jsonData))
		}
	}

	logEntry.WriteString("\n" + strings.Repeat("-", 80) + "\n")

	// 写入文件
	cb.logFile.WriteString(logEntry.String())
	cb.logFile.Sync()
}

// OnStart 开始执行时的回调
func (cb *JavaAnalyzerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	cb.startTime = time.Now()
	cb.writeLog("START", fmt.Sprintf("开始执行: %s", info.Name), map[string]interface{}{
		"component":  info.Component,
		"type":       info.Type,
		"name":       info.Name,
		"input":      input,
		"start_time": cb.startTime.Format(time.RFC3339),
	})
	return ctx
}

// OnEnd 执行结束时的回调
func (cb *JavaAnalyzerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	duration := time.Since(cb.startTime)
	cb.writeLog("END", fmt.Sprintf("执行完成: %s (耗时: %v)", info.Name, duration), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"output":    output,
		"duration":  duration.String(),
		"end_time":  time.Now().Format(time.RFC3339),
	})
	return ctx
}

// OnError 发生错误时的回调
func (cb *JavaAnalyzerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	duration := time.Since(cb.startTime)
	cb.writeLog("ERROR", fmt.Sprintf("执行出错: %s (耗时: %v)", info.Name, duration), map[string]interface{}{
		"component":  info.Component,
		"type":       info.Type,
		"name":       info.Name,
		"error":      err.Error(),
		"duration":   duration.String(),
		"error_time": time.Now().Format(time.RFC3339),
	})
	return ctx
}

// OnEndWithStreamOutput 流式输出结束时的回调
func (cb *JavaAnalyzerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	// 只记录 React Agent 的主要输出，避免重复记录
	var graphInfoName = react.GraphName

	go func() {
		defer func() {
			if err := recover(); err != nil {
				cb.writeLog("PANIC", fmt.Sprintf("流式输出处理panic: %v", err), nil)
			}
		}()

		defer output.Close()

		var streamContent strings.Builder
		streamContent.WriteString("流式输出内容:\n")

		stepCount := 0
		for {
			frame, err := output.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				cb.writeLog("ERROR", fmt.Sprintf("流式输出读取错误: %v", err), nil)
				return
			}

			stepCount++
			streamContent.WriteString(fmt.Sprintf("步骤 %d:\n", stepCount))

			// 序列化输出内容
			jsonData, err := json.MarshalIndent(frame, "", "  ")
			if err != nil {
				streamContent.WriteString(fmt.Sprintf("序列化失败: %v\n", err))
			} else {
				streamContent.WriteString(string(jsonData))
				streamContent.WriteString("\n")
			}

			// 只记录 React Agent 的主要输出
			if info.Name == graphInfoName {
				cb.writeLog("STREAM", fmt.Sprintf("React Agent 输出步骤 %d", stepCount), frame)
			}
		}

		// 记录完整的流式输出摘要
		cb.writeLog("STREAM_SUMMARY", fmt.Sprintf("流式输出完成: %s (共 %d 步)", info.Name, stepCount), map[string]interface{}{
			"component":  info.Component,
			"type":       info.Type,
			"name":       info.Name,
			"step_count": stepCount,
			"content":    streamContent.String(),
		})
	}()

	return ctx
}

// OnStartWithStreamInput 流式输入开始时的回调
func (cb *JavaAnalyzerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {

	cb.writeLog("STREAM_INPUT", fmt.Sprintf("开始流式输入: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
	})

	defer input.Close()
	return ctx
}

// OnToolStart 工具开始执行时的回调
func (cb *JavaAnalyzerCallback) OnToolStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	cb.writeLog("TOOL_START", fmt.Sprintf("工具开始执行: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"input":     input,
	})
	return ctx
}

// OnToolEnd 工具执行结束时的回调
func (cb *JavaAnalyzerCallback) OnToolEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	cb.writeLog("TOOL_END", fmt.Sprintf("工具执行完成: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"output":    output,
	})
	return ctx
}

// OnToolError 工具执行出错时的回调
func (cb *JavaAnalyzerCallback) OnToolError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	cb.writeLog("TOOL_ERROR", fmt.Sprintf("工具执行出错: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"error":     err.Error(),
	})
	return ctx
}

// OnLLMStart LLM开始调用时的回调
func (cb *JavaAnalyzerCallback) OnLLMStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	cb.writeLog("LLM_START", fmt.Sprintf("LLM开始调用: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"input":     input,
	})
	return ctx
}

// OnLLMEnd LLM调用结束时的回调
func (cb *JavaAnalyzerCallback) OnLLMEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	cb.writeLog("LLM_END", fmt.Sprintf("LLM调用完成: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"output":    output,
	})
	return ctx
}

// OnLLMError LLM调用出错时的回调
func (cb *JavaAnalyzerCallback) OnLLMError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	cb.writeLog("LLM_ERROR", fmt.Sprintf("LLM调用出错: %s", info.Name), map[string]interface{}{
		"component": info.Component,
		"type":      info.Type,
		"name":      info.Name,
		"error":     err.Error(),
	})
	return ctx
}

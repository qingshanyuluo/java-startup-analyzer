package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Formatter 格式化器接口
type Formatter interface {
	Format(result *AnalysisResult, output io.Writer) error
}

// TextFormatter 文本格式化器
type TextFormatter struct{}

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

// MarkdownFormatter Markdown格式化器
type MarkdownFormatter struct{}

// NewFormatter 创建格式化器
func NewFormatter(format string) Formatter {
	switch strings.ToLower(format) {
	case "json":
		return &JSONFormatter{}
	case "markdown", "md":
		return &MarkdownFormatter{}
	default:
		return &TextFormatter{}
	}
}

// Format 文本格式化
func (f *TextFormatter) Format(result *AnalysisResult, output io.Writer) error {
	var builder strings.Builder

	// 标题
	builder.WriteString("=== Java启动失败分析报告 ===\n\n")

	// 基本信息
	builder.WriteString("📊 基本信息:\n")
	builder.WriteString(fmt.Sprintf("  分析时间: %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("  日志大小: %d 字符\n", result.LogSize))
	builder.WriteString(fmt.Sprintf("  分析耗时: %v\n", result.AnalysisTime))
	builder.WriteString(fmt.Sprintf("  置信度: %.1f%%\n\n", result.Confidence*100))

	// 分析结果
	builder.WriteString("🔍 分析结果:\n")
	builder.WriteString(fmt.Sprintf("  状态: %s\n", getStatusEmoji(result.Status)))
	builder.WriteString(fmt.Sprintf("  错误类型: %s\n", result.ErrorType))
	builder.WriteString(fmt.Sprintf("  错误消息: %s\n", result.ErrorMessage))
	builder.WriteString(fmt.Sprintf("  根本原因: %s\n\n", result.RootCause))

	// 摘要
	if result.Summary != "" {
		builder.WriteString("📝 摘要:\n")
		builder.WriteString(fmt.Sprintf("  %s\n\n", result.Summary))
	}

	// 解决建议
	if len(result.Suggestions) > 0 {
		builder.WriteString("💡 解决建议:\n")
		for i, suggestion := range result.Suggestions {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
		builder.WriteString("\n")
	}

	// 相关错误
	if len(result.RelatedErrors) > 0 {
		builder.WriteString("🔗 相关错误:\n")
		for i, error := range result.RelatedErrors {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, error))
		}
		builder.WriteString("\n")
	}

	// 详细信息
	if result.Details != "" {
		builder.WriteString("📋 详细信息:\n")
		lines := strings.Split(result.Details, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				builder.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
		builder.WriteString("\n")
	}

	// 元数据
	if len(result.Metadata) > 0 {
		builder.WriteString("🏷️  元数据:\n")
		for key, value := range result.Metadata {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	_, err := output.Write([]byte(builder.String()))
	return err
}

// Format JSON格式化
func (f *JSONFormatter) Format(result *AnalysisResult, output io.Writer) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// Format Markdown格式化
func (f *MarkdownFormatter) Format(result *AnalysisResult, output io.Writer) error {
	var builder strings.Builder

	// 标题
	builder.WriteString("# Java启动失败分析报告\n\n")

	// 基本信息表格
	builder.WriteString("## 📊 基本信息\n\n")
	builder.WriteString("| 项目 | 值 |\n")
	builder.WriteString("|------|-----|\n")
	builder.WriteString(fmt.Sprintf("| 分析时间 | %s |\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("| 日志大小 | %d 字符 |\n", result.LogSize))
	builder.WriteString(fmt.Sprintf("| 分析耗时 | %v |\n", result.AnalysisTime))
	builder.WriteString(fmt.Sprintf("| 置信度 | %.1f%% |\n\n", result.Confidence*100))

	// 分析结果
	builder.WriteString("## 🔍 分析结果\n\n")
	builder.WriteString(fmt.Sprintf("- **状态**: %s %s\n", getStatusEmoji(result.Status), result.Status))
	builder.WriteString(fmt.Sprintf("- **错误类型**: %s\n", result.ErrorType))
	builder.WriteString(fmt.Sprintf("- **错误消息**: %s\n", result.ErrorMessage))
	builder.WriteString(fmt.Sprintf("- **根本原因**: %s\n\n", result.RootCause))

	// 摘要
	if result.Summary != "" {
		builder.WriteString("## 📝 摘要\n\n")
		builder.WriteString(fmt.Sprintf("%s\n\n", result.Summary))
	}

	// 解决建议
	if len(result.Suggestions) > 0 {
		builder.WriteString("## 💡 解决建议\n\n")
		for i, suggestion := range result.Suggestions {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
		builder.WriteString("\n")
	}

	// 相关错误
	if len(result.RelatedErrors) > 0 {
		builder.WriteString("## 🔗 相关错误\n\n")
		for i, error := range result.RelatedErrors {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, error))
		}
		builder.WriteString("\n")
	}

	// 详细信息
	if result.Details != "" {
		builder.WriteString("## 📋 详细信息\n\n")
		builder.WriteString("```\n")
		builder.WriteString(result.Details)
		builder.WriteString("\n```\n\n")
	}

	// 元数据
	if len(result.Metadata) > 0 {
		builder.WriteString("## 🏷️ 元数据\n\n")
		builder.WriteString("| 键 | 值 |\n")
		builder.WriteString("|----|----|\n")
		for key, value := range result.Metadata {
			builder.WriteString(fmt.Sprintf("| %s | %s |\n", key, value))
		}
	}

	_, err := output.Write([]byte(builder.String()))
	return err
}

// getStatusEmoji 获取状态对应的表情符号
func getStatusEmoji(status string) string {
	switch strings.ToLower(status) {
	case "success":
		return "✅"
	case "failure":
		return "❌"
	case "warning":
		return "⚠️"
	default:
		return "❓"
	}
}

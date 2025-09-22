package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/user/java-startup-analyzer/internal/llm"
	"github.com/user/java-startup-analyzer/pkg/logparser"
)

// JavaAnalyzer Java启动分析器
type JavaAnalyzer struct {
	config *Config
	chain  compose.Runnable[map[string]any, *schema.Message]
}

// NewJavaAnalyzer 创建新的Java分析器
func NewJavaAnalyzer(config *Config) (*JavaAnalyzer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建LLM客户端
	llmClient, err := llm.NewClient(config.Model, config.ModelName, config.APIKey, config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	// 创建分析链
	chain, err := createAnalysisChain(llmClient)
	if err != nil {
		return nil, fmt.Errorf("创建分析链失败: %w", err)
	}

	return &JavaAnalyzer{
		config: config,
		chain:  chain,
	}, nil
}

// Analyze 分析Java启动日志
func (ja *JavaAnalyzer) Analyze(logContent string) (*AnalysisResult, error) {
	startTime := time.Now()

	// 预处理日志
	processedLog, err := ja.preprocessLog(logContent)
	if err != nil {
		return nil, fmt.Errorf("预处理日志失败: %w", err)
	}

	// 使用LLM分析
	ctx := context.Background()
	input := map[string]any{
		"log_content": processedLog,
		"log_size":    len(logContent),
	}

	result, err := ja.chain.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("LLM分析失败: %w", err)
	}

	// 解析结果
	analysisResult, err := ja.parseAnalysisResult(result, logContent, time.Since(startTime))
	if err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	return analysisResult, nil
}

// preprocessLog 预处理日志内容
func (ja *JavaAnalyzer) preprocessLog(logContent string) (string, error) {
	// 使用日志解析器提取关键信息
	parser := logparser.NewJavaLogParser()
	parsedLog, err := parser.Parse(logContent)
	if err != nil {
		// 如果解析失败，返回原始内容
		if ja.config.Verbose {
			fmt.Printf("警告: 日志解析失败，使用原始内容: %v\n", err)
		}
		return logContent, nil
	}

	// 构建简化的日志内容
	var builder strings.Builder
	builder.WriteString("=== Java启动日志分析 ===\n\n")

	if parsedLog.StartupTime != "" {
		builder.WriteString(fmt.Sprintf("启动时间: %s\n", parsedLog.StartupTime))
	}

	if parsedLog.JavaVersion != "" {
		builder.WriteString(fmt.Sprintf("Java版本: %s\n", parsedLog.JavaVersion))
	}

	if parsedLog.MainClass != "" {
		builder.WriteString(fmt.Sprintf("主类: %s\n", parsedLog.MainClass))
	}

	builder.WriteString("\n=== 错误信息 ===\n")
	for _, error := range parsedLog.Errors {
		builder.WriteString(fmt.Sprintf("- %s\n", error))
	}

	builder.WriteString("\n=== 异常堆栈 ===\n")
	for _, stack := range parsedLog.StackTraces {
		builder.WriteString(fmt.Sprintf("%s\n", stack))
	}

	builder.WriteString("\n=== 完整日志 ===\n")
	builder.WriteString(logContent)

	return builder.String(), nil
}

// parseAnalysisResult 解析LLM分析结果
func (ja *JavaAnalyzer) parseAnalysisResult(result *schema.Message, originalLog string, analysisTime time.Duration) (*AnalysisResult, error) {
	analysisResult := NewAnalysisResult()
	analysisResult.LogSize = len(originalLog)
	analysisResult.AnalysisTime = analysisTime

	// 解析LLM返回的内容
	content := result.Content
	lines := strings.Split(content, "\n")

	var currentSection string
	var details strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 识别不同的部分
		if strings.HasPrefix(line, "状态:") || strings.HasPrefix(line, "Status:") {
			analysisResult.Status = strings.TrimSpace(strings.TrimPrefix(line, "状态:"))
			analysisResult.Status = strings.TrimSpace(strings.TrimPrefix(analysisResult.Status, "Status:"))
		} else if strings.HasPrefix(line, "错误类型:") || strings.HasPrefix(line, "Error Type:") {
			analysisResult.ErrorType = strings.TrimSpace(strings.TrimPrefix(line, "错误类型:"))
			analysisResult.ErrorType = strings.TrimSpace(strings.TrimPrefix(analysisResult.ErrorType, "Error Type:"))
		} else if strings.HasPrefix(line, "错误消息:") || strings.HasPrefix(line, "Error Message:") {
			analysisResult.ErrorMessage = strings.TrimSpace(strings.TrimPrefix(line, "错误消息:"))
			analysisResult.ErrorMessage = strings.TrimSpace(strings.TrimPrefix(analysisResult.ErrorMessage, "Error Message:"))
		} else if strings.HasPrefix(line, "根本原因:") || strings.HasPrefix(line, "Root Cause:") {
			analysisResult.RootCause = strings.TrimSpace(strings.TrimPrefix(line, "根本原因:"))
			analysisResult.RootCause = strings.TrimSpace(strings.TrimPrefix(analysisResult.RootCause, "Root Cause:"))
		} else if strings.HasPrefix(line, "建议:") || strings.HasPrefix(line, "Suggestions:") {
			currentSection = "suggestions"
		} else if strings.HasPrefix(line, "摘要:") || strings.HasPrefix(line, "Summary:") {
			analysisResult.Summary = strings.TrimSpace(strings.TrimPrefix(line, "摘要:"))
			analysisResult.Summary = strings.TrimSpace(strings.TrimPrefix(analysisResult.Summary, "Summary:"))
		} else if strings.HasPrefix(line, "置信度:") || strings.HasPrefix(line, "Confidence:") {
			// 解析置信度
			confidenceStr := strings.TrimSpace(strings.TrimPrefix(line, "置信度:"))
			confidenceStr = strings.TrimSpace(strings.TrimPrefix(confidenceStr, "Confidence:"))
			if confidenceStr != "" {
				// 简单的置信度解析，实际应用中可能需要更复杂的解析
				if strings.Contains(confidenceStr, "高") || strings.Contains(confidenceStr, "high") {
					analysisResult.Confidence = 0.8
				} else if strings.Contains(confidenceStr, "中") || strings.Contains(confidenceStr, "medium") {
					analysisResult.Confidence = 0.6
				} else if strings.Contains(confidenceStr, "低") || strings.Contains(confidenceStr, "low") {
					analysisResult.Confidence = 0.3
				}
			}
		} else if currentSection == "suggestions" && strings.HasPrefix(line, "-") {
			suggestion := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			analysisResult.Suggestions = append(analysisResult.Suggestions, suggestion)
		} else {
			// 添加到详细信息
			details.WriteString(line)
			details.WriteString("\n")
		}
	}

	analysisResult.Details = details.String()

	// 设置默认值
	if analysisResult.Status == "" {
		analysisResult.Status = "unknown"
	}
	if analysisResult.Confidence == 0.0 {
		analysisResult.Confidence = 0.5 // 默认中等置信度
	}

	return analysisResult, nil
}

// createAnalysisChain 创建分析链
func createAnalysisChain(llmClient *llm.Client) (compose.Runnable[map[string]any, *schema.Message], error) {
	// 创建系统提示模板
	systemPrompt := `你是一个专业的Java应用程序启动问题诊断专家。你的任务是分析Java应用程序的启动日志，识别启动失败的原因并提供专业的解决建议。

请按照以下格式输出分析结果：

状态: [success/failure/warning]
错误类型: [具体的错误类型，如ClassNotFoundException, OutOfMemoryError等]
错误消息: [主要的错误消息]
根本原因: [启动失败的根本原因分析]
摘要: [简要的问题描述和影响]
置信度: [高/中/低]

建议:
- [具体的解决建议1]
- [具体的解决建议2]
- [具体的解决建议3]

请重点关注以下常见的Java启动问题：
1. 类路径问题 (ClassNotFoundException, NoClassDefFoundError)
2. 内存问题 (OutOfMemoryError)
3. 端口占用问题
4. 配置文件问题
5. 依赖冲突问题
6. 权限问题
7. 环境变量问题

请提供具体、可操作的解决建议。`

	// 创建用户提示模板
	userPrompt := `请分析以下Java启动日志：

日志大小: {{.log_size}} 字符

日志内容:
{{.log_content}}

请提供详细的分析结果。`

	// 创建链
	chain, err := compose.NewChain[map[string]any, *schema.Message]().
		AppendChatTemplate(createChatTemplate(systemPrompt, userPrompt)).
		AppendChatModel(llmClient.GetChatModel()).
		Compile(context.Background())

	if err != nil {
		return nil, fmt.Errorf("编译分析链失败: %w", err)
	}

	return chain, nil
}

// createChatTemplate 创建聊天模板
func createChatTemplate(systemPrompt, userPrompt string) prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		&schema.Message{
			Role:    schema.System,
			Content: systemPrompt,
		},
		&schema.Message{
			Role:    schema.User,
			Content: userPrompt,
		},
	)
}

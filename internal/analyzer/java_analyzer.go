package analyzer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/user/java-startup-analyzer/internal/llm"
)

// JavaAnalyzer Java启动分析器
type JavaAnalyzer struct {
	config *Config
	agent  *react.Agent
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

	// 创建分析代理
	agent, err := createAnalysisAgent(llmClient)
	if err != nil {
		return nil, fmt.Errorf("创建分析代理失败: %w", err)
	}

	return &JavaAnalyzer{
		config: config,
		agent:  agent,
	}, nil
}

func (ja *JavaAnalyzer) Chat(ctx context.Context, input map[string]any) (*schema.Message, error) {
	// Convert input map to messages
	var messages []*schema.Message

	// Add user message with the input content
	if logContent, ok := input["log_content"].(string); ok {
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: logContent,
		}
		messages = append(messages, userMessage)
	}

	return ja.agent.Generate(ctx, messages)
}

// ChatStream 流式聊天方法
func (ja *JavaAnalyzer) ChatStream(ctx context.Context, input map[string]any) (*schema.StreamReader[*schema.Message], error) {
	// Convert input map to messages
	var messages []*schema.Message

	// Add user message with the input content
	if logContent, ok := input["log_content"].(string); ok {
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: logContent,
		}
		messages = append(messages, userMessage)
	}

	return ja.agent.Stream(ctx, messages)
}

// createAnalysisAgent 创建分析代理
func createAnalysisAgent(llmClient *llm.Client) (*react.Agent, error) {
	// 创建系统提示模板
	systemPrompt := `你是一个专业的Java应用程序启动问题诊断专家。你的任务是分析Java应用程序的启动日志，识别启动失败的原因并提供专业的解决建议。

你还可以使用工具来帮助你分析。`

	// 创建代理配置
	config := &react.AgentConfig{
		Model:           llmClient.GetChatModel(),
		MessageModifier: react.NewPersonaModifier(systemPrompt),
	}

	// 创建代理
	reactAgent, err := react.NewAgent(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("创建ReAct代理失败: %w", err)
	}

	return reactAgent, nil
}

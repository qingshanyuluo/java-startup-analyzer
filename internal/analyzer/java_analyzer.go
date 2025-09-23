package analyzer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/user/java-startup-analyzer/internal/llm"
	"github.com/user/java-startup-analyzer/internal/tools"
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

	// 绑定工具到LLM模型（ReAct Agent 需要模型支持工具调用）
	analyzerTools := tools.GetAnalyzerTools()
	if err := llmClient.BindTools(analyzerTools); err != nil {
		return nil, fmt.Errorf("绑定工具失败: %w", err)
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

	// Add user message with file path instead of log content
	if logPath, ok := input["log_path"].(string); ok {
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: fmt.Sprintf("请分析这个Java应用日志文件: %s", logPath),
		}
		messages = append(messages, userMessage)
	} else if _, ok := input["log_content"].(string); ok {
		// 如果提供了日志内容，指导用户使用文件路径
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: "请提供日志文件的路径，我将使用工具来读取和分析日志内容。",
		}
		messages = append(messages, userMessage)
	} else {
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: "请提供Java应用日志文件的路径进行分析。",
		}
		messages = append(messages, userMessage)
	}

	return ja.agent.Generate(ctx, messages)
}

// ChatStream 流式聊天方法
func (ja *JavaAnalyzer) ChatStream(ctx context.Context, input map[string]any) (*schema.StreamReader[*schema.Message], error) {
	// Convert input map to messages
	var messages []*schema.Message

	// Add user message with file path instead of log content
	if logPath, ok := input["log_path"].(string); ok {
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: fmt.Sprintf("请分析这个Java应用日志文件: %s", logPath),
		}
		messages = append(messages, userMessage)
	} else if _, ok := input["log_content"].(string); ok {
		// 如果提供了日志内容，指导用户使用文件路径
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: "请提供日志文件的路径，我将使用工具来读取和分析日志内容。",
		}
		messages = append(messages, userMessage)
	} else {
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: "请提供Java应用日志文件的路径进行分析。",
		}
		messages = append(messages, userMessage)
	}

	return ja.agent.Stream(ctx, messages)
}

// createAnalysisAgent 创建分析代理
func createAnalysisAgent(llmClient *llm.Client) (*react.Agent, error) {
	// 创建系统提示模板
	systemPrompt := `你是一个专业的Java应用程序启动问题诊断专家。你的任务是分析Java应用程序的启动日志，识别启动失败的原因并提供专业的解决建议。

你可以使用以下工具：
- tail: 读取日志文件的最后N行内容

分析流程：
1. 当用户提供日志文件路径时，使用tail工具读取日志内容
2. 分析日志中的错误信息、异常堆栈和警告
3. 识别常见的Java启动问题，如：
   - OutOfMemoryError (内存不足)
   - ClassNotFoundException (类未找到)
   - NoSuchMethodError (方法未找到)
   - Connection refused (连接被拒绝)
   - Port already in use (端口被占用)
   - 配置错误
   - 依赖问题
4. 提供详细的诊断结果和具体的解决方案

请始终使用tail工具来读取日志文件，不要要求用户直接提供日志内容。`

	// 创建代理配置
	config := &react.AgentConfig{
		ToolCallingModel: llmClient.GetChatModel().(model.ToolCallingChatModel),
		MessageModifier:  react.NewPersonaModifier(systemPrompt),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				tools.TailTool,
			},
		},
	}

	// 创建代理
	reactAgent, err := react.NewAgent(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("创建ReAct代理失败: %w", err)
	}

	return reactAgent, nil
}

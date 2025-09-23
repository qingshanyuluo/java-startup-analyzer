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
	config   *Config
	agent    *react.Agent
	messages []*schema.Message // 维护对话历史
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
	// 添加新的用户消息到对话历史
	var userMessage *schema.Message

	// Add user message with file path instead of log content
	if logPath, ok := input["log_path"].(string); ok {
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: fmt.Sprintf("请分析这个Java应用日志文件: %s", logPath),
		}
	} else if userInput, ok := input["input"].(string); ok {
		// 处理用户输入（继续聊天）
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: userInput,
		}
	} else if _, ok := input["log_content"].(string); ok {
		// 如果提供了日志内容，指导用户使用文件路径
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: "请提供日志文件的路径，我将使用工具来读取和分析日志内容。",
		}
	} else {
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: "请提供Java应用日志文件的路径进行分析。",
		}
	}

	// 将用户消息添加到对话历史
	ja.messages = append(ja.messages, userMessage)

	// 创建包含系统消息的完整消息列表
	messagesWithSystem := make([]*schema.Message, 0, len(ja.messages)+1)
	messagesWithSystem = append(messagesWithSystem, &schema.Message{
		Role:    schema.System,
		Content: systemPrompt,
	})
	messagesWithSystem = append(messagesWithSystem, ja.messages...)

	// 使用完整的对话历史调用 agent
	response, err := ja.agent.Generate(ctx, messagesWithSystem)
	if err != nil {
		return nil, err
	}

	// 将助手的回复也添加到对话历史
	ja.messages = append(ja.messages, response)

	return response, nil
}

// ChatStream 流式聊天方法
func (ja *JavaAnalyzer) ChatStream(ctx context.Context, input map[string]any) (*schema.StreamReader[*schema.Message], error) {
	// 添加新的用户消息到对话历史
	var userMessage *schema.Message

	// Add user message with file path instead of log content
	if logPath, ok := input["log_path"].(string); ok {
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: fmt.Sprintf("请分析这个Java应用日志文件: %s", logPath),
		}
	} else if userInput, ok := input["input"].(string); ok {
		// 处理用户输入（继续聊天）
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: userInput,
		}
	} else if _, ok := input["log_content"].(string); ok {
		// 如果提供了日志内容，指导用户使用文件路径
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: "请提供日志文件的路径，我将使用工具来读取和分析日志内容。",
		}
	} else {
		userMessage = &schema.Message{
			Role:    schema.User,
			Content: "请提供Java应用日志文件的路径进行分析。",
		}
	}

	// 将用户消息添加到对话历史
	ja.messages = append(ja.messages, userMessage)

	// 创建包含系统消息的完整消息列表
	messagesWithSystem := make([]*schema.Message, 0, len(ja.messages)+1)
	messagesWithSystem = append(messagesWithSystem, &schema.Message{
		Role:    schema.System,
		Content: systemPrompt,
	})
	messagesWithSystem = append(messagesWithSystem, ja.messages...)

	// 使用完整的对话历史调用 agent
	streamReader, err := ja.agent.Stream(ctx, messagesWithSystem)
	if err != nil {
		return nil, err
	}

	// 注意：流式响应中，我们需要在流式处理完成后将助手的回复添加到对话历史
	// 这需要在 UI 层处理流式响应时完成

	return streamReader, nil
}

// AddAssistantMessage 将助手的回复添加到对话历史（用于流式响应完成后）
func (ja *JavaAnalyzer) AddAssistantMessage(content string) {
	assistantMessage := &schema.Message{
		Role:    schema.Assistant,
		Content: content,
	}
	ja.messages = append(ja.messages, assistantMessage)
}

// 系统提示模板
const systemPrompt = `你是一个专业的Java应用程序启动问题诊断专家。你的任务是分析Java应用程序的启动日志，识别启动失败的原因并提供专业的解决建议。

你可以使用以下工具：
- read_file: 读取指定文件的内容，支持分页读取大文件和反向读取
- search_file_content: 在目录中搜索正则表达式模式，用于查找特定的错误信息或配置问题

## 工具使用最佳实践：

### 1. 初始日志分析（推荐方式）
- 首先使用：reverse=true, limit=100
- 这会读取日志文件的最后100行，通常包含最新的错误信息
- 示例：{"absolute_path": "/path/to/log", "reverse": true, "limit": 100}

### 2. 分页读取策略
- 如果需要更多内容，使用offset参数继续读取
- 反向读取：reverse=true, offset=100, limit=100 （读取倒数第101-200行）
- 正向读取：offset=0, limit=100 （读取前100行）

### 3. 搜索工具使用策略
- 当多次读取日志后仍未找到明确错误原因时，使用search_file_content工具
- 搜索常见的Java错误模式：
  - "Exception" - 查找所有异常
  - "Error" - 查找所有错误
  - "OutOfMemoryError" - 内存不足错误
  - "ClassNotFoundException" - 类未找到错误
  - "NoSuchMethodError" - 方法未找到错误
  - "Connection refused" - 连接被拒绝
  - "Port.*already in use" - 端口被占用
  - "Configuration.*error" - 配置错误
  - "finish.*error" - 启动完成时的错误
  - "startup.*failed" - 启动失败
  - "application.*failed" - 应用启动失败
  - "failed.*to.*start" - 启动失败
  - "shutdown.*error" - 关闭错误
  - "timeout" - 超时错误
  - "deadlock" - 死锁
- 示例：{"pattern": "finish.*error", "include": "*.log"}

### 4. 参数说明
- read_file工具：
  - absolute_path: 必须提供绝对路径
  - reverse: true=从末尾开始读取（推荐用于日志分析）
  - limit: 建议初始使用100行，避免一次性读取过多内容
  - offset: 0-based行号，reverse=true时从末尾计算
- search_file_content工具：
  - pattern: 正则表达式模式（必需）
  - path: 搜索目录路径（可选，默认为当前目录）
  - include: 文件过滤模式（可选，如"*.log", "*.java"）

## 分析流程：
1. 当用户提供日志文件路径时，立即使用read_file工具读取最后100行（必须至少查看100行）
2. 分析日志中的错误信息、异常堆栈和警告
3. 如果100行不够，根据分析结果决定是否需要读取更多内容（最多200行）
4. 如果多次读取后仍未找到明确原因，使用search_file_content工具搜索相关错误模式
5. 必须进行关键词搜索，包括但不限于：
   - "finish.*error" - 启动完成时的错误
   - "Exception" - 所有异常
   - "Error" - 所有错误
   - "failed.*to.*start" - 启动失败
   - "startup.*failed" - 启动失败
6. 识别常见的Java启动问题，如：
   - OutOfMemoryError (内存不足)
   - ClassNotFoundException (类未找到)
   - NoSuchMethodError (方法未找到)
   - Connection refused (连接被拒绝)
   - Port already in use (端口被占用)
   - 配置错误
   - 依赖问题
   - 启动完成时的错误
   - 超时问题
   - 死锁问题
7. 提供详细的诊断结果和具体的解决方案

## 重要提醒：
- 始终使用read_file工具来读取日志文件，不要要求用户直接提供日志内容
- 必须至少查看最后100行，优先使用reverse=true读取最后100行，因为错误通常出现在日志末尾
- 如果文件很大，分页读取而不是一次性读取全部内容（最多200行）
- 当多次读取后仍未找到问题根源时，必须使用search_file_content工具进行深度搜索
- 必须搜索"finish.*error"等关键词，进行全面分析
- 搜索工具可以帮助找到分散在多个文件中的相关错误信息
- 分析必须全面，不能遗漏任何可能的错误模式
- 重点关注启动完成时的错误和启动失败的相关信息`

// createAnalysisAgent 创建分析代理
func createAnalysisAgent(llmClient *llm.Client) (*react.Agent, error) {
	// 创建代理配置
	config := &react.AgentConfig{
		ToolCallingModel: llmClient.GetChatModel().(model.ToolCallingChatModel),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				tools.ReadFileTool,
				tools.SearchFileContentTool,
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

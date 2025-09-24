package analyzer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/user/java-startup-analyzer/internal/llm"
	"github.com/user/java-startup-analyzer/internal/tools"
)

// JavaAnalyzer Java启动分析器
type JavaAnalyzer struct {
	config   *Config
	agent    *react.Agent
	callback *JavaAnalyzerCallback
}

// modifyJavaAnalyzerMessages MessageModifier 函数，用于管理历史记录和消息长度限制
func modifyJavaAnalyzerMessages(ctx context.Context, input []*schema.Message) []*schema.Message {
	sum := 0
	maxLimit := 50000 // 单个消息最大长度限制
	maxMessages := 20 // 最大消息数量限制

	// 如果消息数量超过限制，保留最新的消息
	if len(input) > maxMessages {
		input = input[len(input)-maxMessages:]
	}

	for i := range input {
		if input[i] == nil {
			continue
		}
		l := len(input[i].Content)
		if l > maxLimit {
			// 截取消息末尾部分，保留最新的内容
			input[i].Content = input[i].Content[l-maxLimit:]
		}
		sum += len(input[i].Content)
	}

	return input
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

	// 创建回调处理器
	callback, err := NewJavaAnalyzerCallback(config.LogDir)
	if err != nil {
		return nil, fmt.Errorf("创建回调处理器失败: %w", err)
	}

	// 创建分析代理
	agent, err := createAnalysisAgent(llmClient, callback)
	if err != nil {
		callback.Close() // 清理资源
		return nil, fmt.Errorf("创建分析代理失败: %w", err)
	}

	return &JavaAnalyzer{
		config:   config,
		agent:    agent,
		callback: callback,
	}, nil
}

// ChatStream 流式聊天方法
func (ja *JavaAnalyzer) ChatStream(ctx context.Context, input map[string]any) (*schema.StreamReader[*schema.Message], error) {
	// 创建用户消息
	var userMessage *schema.Message

	// 根据输入类型创建相应的用户消息
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

	// 创建包含系统消息和用户消息的消息列表
	// MessageModifier 会自动管理历史记录和消息长度
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		userMessage,
	}

	// 使用回调系统记录执行过程
	streamReader, err := ja.agent.Stream(ctx, messages,
		agent.WithComposeOptions(compose.WithCallbacks(ja.callback)))
	if err != nil {
		return nil, err
	}

	return streamReader, nil
}

// GetLogPath 获取当前会话的日志文件路径
func (ja *JavaAnalyzer) GetLogPath() string {
	if ja.callback != nil {
		return ja.callback.GetLogPath()
	}
	return ""
}

// Close 关闭分析器并清理资源
func (ja *JavaAnalyzer) Close() error {
	if ja.callback != nil {
		return ja.callback.Close()
	}
	return nil
}

// 系统提示模板
const systemPrompt = `你是一个专业的Spring Boot应用程序启动问题诊断专家。你的任务是分析Spring Boot应用程序的启动日志，首先判断应用是否启动成功，然后识别启动失败的原因或运行中的问题，并提供专业的解决建议。

你可以使用以下工具：
- read_file: 读取指定文件的内容，支持分页读取大文件和反向读取
- search_file_content: 在目录中搜索正则表达式模式，用于查找特定的错误信息或配置问题

## Spring Boot启动成功判断标准：

### ✅ 启动成功的核心指标
1. **应用启动完成标志**：查找 "Started [ApplicationName] in [time] seconds" 或类似信息
   - 搜索模式：Started.*in.*seconds
   - 示例：Started Application in 24.427 seconds (JVM running for 26.208)

2. **Web服务器启动**：Tomcat/Undertow/Jetty等嵌入式服务器启动成功
   - 搜索模式：Tomcat started on port|Undertow started|Jetty started
   - 示例：Tomcat started on port(s): 8080 (http) with context path ''

3. **数据库连接成功**：数据源连接池启动完成
   - 搜索模式：HikariDataSource.*Start completed|DataSource.*initialized
   - 示例：HikariDataSource - Start completed

4. **健康检查端点**：健康检查服务启动（如果配置了）
   - 搜索模式：Health check.*started|Actuator.*started

### ⚠️ 启动成功但有问题的指标
- 应用启动成功但存在WARN级别的警告
- 依赖冲突（如NoSuchFieldError、NoSuchMethodError）
- 配置问题（如缺少配置项、端口冲突等）
- 服务连接问题（如外部服务不可用）

### ❌ 启动失败的指标
- 应用进程异常退出
- 关键组件初始化失败
- 致命错误（FATAL、ERROR级别）
- 启动超时或卡死

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
- 搜索Spring Boot特定的模式：
  - "Started.*in.*seconds" - 应用启动完成标志
  - "Tomcat started|Undertow started|Jetty started" - Web服务器启动
  - "HikariDataSource.*Start completed" - 数据库连接成功
  - "Exception" - 查找所有异常
  - "Error" - 查找所有错误
  - "OutOfMemoryError" - 内存不足错误
  - "ClassNotFoundException" - 类未找到错误
  - "NoSuchMethodError|NoSuchFieldError" - 方法/字段未找到错误
  - "Connection refused" - 连接被拒绝
  - "Port.*already in use" - 端口被占用
  - "Configuration.*error" - 配置错误
  - "startup.*failed" - 启动失败
  - "application.*failed" - 应用启动失败
  - "failed.*to.*start" - 启动失败
  - "shutdown.*error" - 关闭错误
  - "timeout" - 超时错误
  - "deadlock" - 死锁
  - "WARN" - 警告信息
  - "ERROR" - 错误信息
- 示例：{"pattern": "Started.*in.*seconds", "include": "*.log"}

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

## 分析流程（必须执行多步分析）：
1. **第一步**：使用read_file工具读取最后100行（必须至少查看100行）
2. **第二步**：判断Spring Boot应用是否启动成功
   - 搜索"Started.*in.*seconds"确认启动完成
   - 检查Web服务器是否启动
   - 验证数据库连接是否成功
3. **第三步**：如果100行不够，根据分析结果决定是否需要读取更多内容（最多200行）
4. **第四步**：**必须**使用search_file_content工具搜索相关错误模式，即使read_file已经找到了一些信息
5. **第五步**：必须进行关键词搜索，包括但不限于：
   - "Started.*in.*seconds" - 启动完成标志
   - "Tomcat started|Undertow started" - Web服务器启动
   - "Exception" - 所有异常
   - "Error" - 所有错误
   - "WARN" - 警告信息
   - "failed.*to.*start" - 启动失败
   - "startup.*failed" - 启动失败
6. **第六步**：识别常见的Spring Boot启动问题，如：
   - 启动成功但有警告（依赖冲突、配置问题等）
   - OutOfMemoryError (内存不足)
   - ClassNotFoundException (类未找到)
   - NoSuchMethodError/NoSuchFieldError (方法/字段未找到)
   - Connection refused (连接被拒绝)
   - Port already in use (端口被占用)
   - 配置错误（如Druid连接池配置问题）
   - 依赖问题（如版本冲突）
   - 启动完成时的错误
   - 超时问题
   - 死锁问题
7. **第七步**：提供详细的诊断结果和具体的解决方案
   - 明确说明应用是否启动成功
   - 如果启动成功，列出所有警告和问题
   - 如果启动失败，指出失败原因
   - 提供具体的修复建议

**重要**：你必须执行多步分析，不能仅通过一次read_file就得出结论。必须结合read_file和search_file_content两个工具的结果进行综合分析。

## 重要提醒：
- 始终使用read_file工具来读取日志文件，不要要求用户直接提供日志内容
- 必须至少查看最后100行，优先使用reverse=true读取最后100行，因为错误通常出现在日志末尾
- 如果文件很大，分页读取而不是一次性读取全部内容（最多200行）
- **必须使用search_file_content工具进行深度搜索，这是分析流程的必需步骤**
- 必须搜索"Started.*in.*seconds"等关键词，进行全面分析
- 搜索工具可以帮助找到分散在多个文件中的相关错误信息
- 分析必须全面，不能遗漏任何可能的错误模式
- 重点关注Spring Boot启动成功标志和启动失败的相关信息
- **不要仅通过一次工具调用就得出结论，必须进行多步分析**
- 对于启动成功但有问题的应用，要详细分析所有警告和错误信息`

// createAnalysisAgent 创建分析代理
func createAnalysisAgent(llmClient *llm.Client, callback *JavaAnalyzerCallback) (*react.Agent, error) {
	// 直接创建代理，参考 react.go 例子的结构
	reactAgent, err := react.NewAgent(context.Background(), &react.AgentConfig{
		MaxStep:          10, // 设置最大步数，允许多次工具调用
		ToolCallingModel: llmClient.GetChatModel().(model.ToolCallingChatModel),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				tools.ReadFileTool,
				tools.SearchFileContentTool,
			},
		},
		MessageModifier: modifyJavaAnalyzerMessages, // 添加消息修改器来管理历史记录
	})
	if err != nil {
		return nil, fmt.Errorf("创建ReAct代理失败: %w", err)
	}

	return reactAgent, nil
}

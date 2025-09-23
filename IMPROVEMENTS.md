# Java 启动分析器改进说明

## 概述

根据 Eino 框架的最佳实践，我们对 Java 启动分析器进行了重大改进，使其能够正确使用 ReAct 模型调用工具进行日志分析。现在分析器通过工具自己读取日志文件，而不是直接接收日志内容。

## 主要改进

### 1. 简化的工具支持

我们只保留了一个核心工具：

- **tail**: 读取日志文件的最后N行内容，让 LLM 自己读取和分析日志

### 2. ReAct Agent 集成

- 使用 Eino 的 ReAct Agent 框架
- 正确配置工具映射
- 支持工具调用和结果处理
- LLM 通过工具主动读取日志文件

### 3. 改进的分析流程

- 不再直接发送日志内容给 LLM
- LLM 根据文件路径使用 tail 工具读取日志
- 更符合 ReAct 模式的工作方式
- 保持了流式响应支持

### 4. UI 层修复

- 修复了 `internal/ui/chat.go` 中的 `autoAnalyze()` 函数
- 现在 UI 层也传递文件路径而不是日志内容
- 确保整个应用都使用 function call 模式

## 代码结构

```
internal/
├── analyzer/
│   └── java_analyzer.go      # 主要的分析器实现
├── llm/
│   ├── client.go            # LLM 客户端
│   └── openai.go            # OpenAI 模型实现
└── tools/
    ├── analyzer_tools.go    # 分析工具定义
    └── tail.go             # 文件尾部读取工具
```

## 使用方法

### 基本使用

```go
// 创建配置
config := &analyzer.Config{
    Model:     "openai",
    ModelName: "gpt-3.5-turbo",
    APIKey:    "your-api-key",
    BaseURL:   "",
}

// 创建分析器
analyzer, err := analyzer.NewJavaAnalyzer(config)
if err != nil {
    log.Fatal(err)
}

// 分析日志文件 - 使用文件路径而不是日志内容
input := map[string]any{
    "log_path": "/path/to/your/java-app.log",
}

response, err := analyzer.Chat(ctx, input)
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Content)
```

### 流式分析

```go
// 流式分析
input := map[string]any{
    "log_path": "/path/to/your/java-app.log",
}

stream, err := analyzer.ChatStream(ctx, input)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err != nil {
        break
    }
    fmt.Print(chunk.Content)
}
```

## 工具功能

### tail 工具

查看文件尾部内容，支持：

- 指定行数（默认50行）
- 自动处理文件读取错误
- 文件存在性检查
- 让 LLM 主动读取日志文件进行分析

## 分析流程

1. **用户提供日志文件路径**：通过 `log_path` 参数
2. **LLM 使用 tail 工具**：自动读取日志文件的最后N行
3. **智能分析**：LLM 分析日志内容，识别问题
4. **提供解决方案**：给出具体的诊断结果和解决建议

## 配置选项

```go
type Config struct {
    Model     string // 模型类型 (openai)
    ModelName string // 模型名称 (gpt-3.5-turbo, gpt-4, etc.)
    APIKey    string // API 密钥
    BaseURL   string // 基础 URL (可选，用于自定义端点)
}
```

## 示例场景

### 1. 内存不足错误分析

```go
// 创建日志文件
logContent := `
java.lang.OutOfMemoryError: Java heap space
	at java.util.Arrays.copyOf(Arrays.java:3210)
	at java.util.ArrayList.grow(ArrayList.java:267)
`
os.WriteFile("error.log", []byte(logContent), 0644)

// 分析日志文件
input := map[string]any{
    "log_path": "error.log",
}
response, err := analyzer.Chat(ctx, input)
```

分析器会：
- 使用 tail 工具读取日志文件
- 识别内存不足问题
- 建议增加 JVM 堆内存
- 提供内存泄漏检查建议

### 2. 类未找到错误分析

```go
// 创建日志文件
logContent := `
java.lang.ClassNotFoundException: com.example.MissingClass
	at java.net.URLClassLoader.findClass(URLClassLoader.java:382)
`
os.WriteFile("class-error.log", []byte(logContent), 0644)

// 分析日志文件
input := map[string]any{
    "log_path": "class-error.log",
}
response, err := analyzer.Chat(ctx, input)
```

分析器会：
- 使用 tail 工具读取日志文件
- 识别类路径问题
- 建议检查依赖配置
- 提供解决方案

## 运行测试

```bash
# 运行简单测试
go run test/simple_test.go

# 运行 UI 修复测试
go run test/test_ui_fix.go

# 运行完整示例
go run examples/usage_example.go
```

## 注意事项

1. **API 密钥**: 请确保设置正确的 OpenAI API 密钥
2. **网络连接**: 需要能够访问 OpenAI API
3. **文件路径**: 使用 `log_path` 参数提供日志文件路径，而不是 `log_content`
4. **工具调用**: ReAct Agent 会自动使用 tail 工具读取日志文件
5. **错误处理**: 所有工具都包含适当的错误处理
6. **文件权限**: 确保分析器有权限读取指定的日志文件

## 未来改进

1. 添加更多分析工具
2. 支持更多 LLM 提供商
3. 增强错误诊断能力
4. 添加性能分析工具

## 相关文档

- [Eino 框架文档](https://www.cloudwego.io/zh/docs/eino/)
- [ReAct Agent 使用手册](https://www.cloudwego.io/zh/docs/eino/quick_start/agent_llm_with_tools/)
- [OpenAI ChatModel 集成](https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/chat_model/chat_model_openai/)

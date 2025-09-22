# Java Startup Analyzer 调试指南

本文档介绍如何在Java Startup Analyzer项目中进行调试。

## 目录
- [1. 基本调试方法](#1-基本调试方法)
- [2. 日志调试](#2-日志调试)
- [3. 命令行调试](#3-命令行调试)
- [4. IDE调试](#4-ide调试)
- [5. 网络调试](#5-网络调试)
- [6. 性能调试](#6-性能调试)
- [7. 常见问题](#7-常见问题)

## 1. 基本调试方法

### 1.1 使用verbose模式

```bash
# 启用详细输出模式
./java-analyzer analyze -f error.log --verbose

# 在聊天模式中启用详细输出
./java-analyzer chat --verbose --api-key YOUR_KEY
```

### 1.2 使用配置文件调试

创建调试配置文件 `debug-config.yaml`：

```yaml
# 调试配置
model: "openai"
api_key: "your-debug-api-key"
base_url: ""
verbose: true

# 调试特定配置
debug:
  log_level: "debug"
  enable_tracing: true
  mock_llm: false  # 设置为true使用模拟LLM
  timeout: 60      # 调试时使用较短超时
```

使用调试配置：
```bash
./java-analyzer analyze -f error.log --config debug-config.yaml
```

## 2. 日志调试

### 2.1 添加调试日志

在代码中添加调试日志：

```go
import (
    "log"
    "os"
)

// 设置调试日志
func init() {
    if os.Getenv("DEBUG") == "true" {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
        log.SetOutput(os.Stderr)
    }
}

// 在关键位置添加日志
func (ja *JavaAnalyzer) Analyze(logContent string) (*Result, error) {
    if ja.config.Verbose {
        log.Printf("开始分析日志，长度: %d", len(logContent))
    }
    
    // 分析逻辑...
    
    if ja.config.Verbose {
        log.Printf("分析完成，结果: %+v", result)
    }
    
    return result, nil
}
```

### 2.2 环境变量调试

```bash
# 设置调试环境变量
export DEBUG=true
export JAVA_ANALYZER_VERBOSE=true
export JAVA_ANALYZER_LOG_LEVEL=debug

# 运行程序
./java-analyzer analyze -f error.log
```

## 3. 命令行调试

### 3.1 使用Go的调试标志

```bash
# 构建调试版本
go build -gcflags="all=-N -l" -o java-analyzer-debug main.go

# 使用race detector检测并发问题
go build -race -o java-analyzer-race main.go
./java-analyzer-race analyze -f error.log

# 使用内存分析
go build -o java-analyzer main.go
GODEBUG=gctrace=1 ./java-analyzer analyze -f error.log
```

### 3.2 使用pprof进行性能分析

```go
import (
    _ "net/http/pprof"
    "net/http"
    "log"
)

func init() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

然后访问 `http://localhost:6060/debug/pprof/` 进行性能分析。

## 4. IDE调试

### 4.1 VS Code调试配置

创建 `.vscode/launch.json`：

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Java Analyzer",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["analyze", "-f", "examples/sample-java-error.log", "--verbose"],
            "env": {
                "DEBUG": "true"
            }
        },
        {
            "name": "Debug Chat Mode",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["chat", "--verbose"],
            "env": {
                "DEBUG": "true",
                "JAVA_ANALYZER_API_KEY": "your-api-key"
            }
        }
    ]
}
```

### 4.2 GoLand调试配置

1. 打开 `Run/Debug Configurations`
2. 创建新的 `Go Build` 配置
3. 设置：
   - **Run kind**: File
   - **Files**: `main.go`
   - **Program arguments**: `analyze -f examples/sample-java-error.log --verbose`
   - **Environment variables**: `DEBUG=true`

## 5. 网络调试

### 5.1 HTTP请求调试

```go
import (
    "net/http"
    "net/http/httputil"
    "log"
)

// 添加HTTP调试中间件
func debugHTTP(req *http.Request) {
    if os.Getenv("DEBUG_HTTP") == "true" {
        dump, err := httputil.DumpRequest(req, true)
        if err != nil {
            log.Printf("HTTP请求调试失败: %v", err)
        } else {
            log.Printf("HTTP请求:\n%s", string(dump))
        }
    }
}
```

### 5.2 使用代理调试

```bash
# 使用mitmproxy调试HTTPS请求
mitmproxy -p 8080

# 设置代理环境变量
export HTTP_PROXY=http://localhost:8080
export HTTPS_PROXY=http://localhost:8080
./java-analyzer analyze -f error.log
```

## 6. 性能调试

### 6.1 内存分析

```bash
# 构建支持内存分析的版本
go build -o java-analyzer main.go

# 运行并生成内存profile
./java-analyzer analyze -f error.log &
PID=$!
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

### 6.2 CPU分析

```bash
# 生成CPU profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile
```

### 6.3 基准测试

```go
// internal/analyzer/analyzer_test.go
func BenchmarkAnalyze(b *testing.B) {
    config := &Config{
        Model:   "openai",
        APIKey:  "test-key",
        Verbose: false,
    }
    
    analyzer, _ := NewJavaAnalyzer(config)
    logContent := "测试日志内容..."
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        analyzer.Analyze(logContent)
    }
}
```

运行基准测试：
```bash
go test -bench=. -benchmem ./internal/analyzer/
```

## 7. 常见问题

### 7.1 聊天模式调试

如果聊天模式出现问题：

```bash
# 检查终端兼容性
echo $TERM

# 使用最小化配置测试
./java-analyzer chat --api-key test --verbose

# 检查Bubble Tea依赖
go mod verify
```

### 7.2 LLM API调试

```bash
# 测试API连接
curl -H "Authorization: Bearer YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     https://api.openai.com/v1/models

# 使用模拟模式测试
export JAVA_ANALYZER_MOCK_LLM=true
./java-analyzer analyze -f error.log
```

### 7.3 配置文件调试

```bash
# 验证配置文件语法
./java-analyzer analyze --config .java-analyzer.yaml --verbose

# 检查配置文件加载
./java-analyzer --config .java-analyzer.yaml --help
```

## 8. 调试工具

### 8.1 内置调试命令

```bash
# 检查系统信息
./java-analyzer version

# 验证配置
./java-analyzer config validate

# 测试LLM连接
./java-analyzer test-connection
```

### 8.2 外部调试工具

- **delve**: Go调试器
  ```bash
  go install github.com/go-delve/delve/cmd/dlv@latest
  dlv debug main.go -- analyze -f error.log
  ```

- **go-trace**: 执行跟踪
  ```bash
  go run main.go analyze -f error.log 2> trace.out
  go tool trace trace.out
  ```

- **go-callvis**: 调用图可视化
  ```bash
  go install github.com/ofthehead/go-callvis@latest
  go-callvis -group pkg,type -focus github.com/user/java-startup-analyzer .
  ```

## 9. 调试最佳实践

1. **分层调试**: 从简单到复杂，逐步调试
2. **日志记录**: 在关键位置添加适当的日志
3. **单元测试**: 为每个组件编写测试
4. **集成测试**: 测试整个流程
5. **性能监控**: 定期检查性能指标
6. **错误处理**: 实现完善的错误处理机制

## 10. 调试脚本

创建 `debug.sh` 脚本：

```bash
#!/bin/bash

echo "🔍 Java Startup Analyzer 调试工具"
echo "================================"

case $1 in
    "build")
        echo "构建调试版本..."
        go build -gcflags="all=-N -l" -o java-analyzer-debug main.go
        ;;
    "test")
        echo "运行测试..."
        go test -v ./...
        ;;
    "race")
        echo "检测竞态条件..."
        go build -race -o java-analyzer-race main.go
        ./java-analyzer-race analyze -f examples/sample-java-error.log
        ;;
    "profile")
        echo "性能分析..."
        go build -o java-analyzer main.go
        ./java-analyzer analyze -f examples/sample-java-error.log &
        sleep 2
        go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile
        ;;
    *)
        echo "用法: $0 {build|test|race|profile}"
        ;;
esac
```

使用方法：
```bash
chmod +x debug.sh
./debug.sh build
./debug.sh test
./debug.sh race
./debug.sh profile
```

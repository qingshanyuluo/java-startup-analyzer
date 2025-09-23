# 项目结构说明

## 目录结构

```
java-startup-analyzer/
├── cmd/                          # 命令行工具
│   ├── chat.go                   # 聊天命令
│   └── root.go                   # 根命令
├── examples/                     # 示例文件
│   ├── chat_history_example.go   # 聊天历史示例
│   ├── conversation_history_example.go  # 对话历史示例
│   ├── eino_openai_example.go    # Eino OpenAI 示例
│   ├── out-of-memory-error.log   # 内存错误日志示例
│   ├── sample-java-error.log     # Java 错误日志示例
│   ├── simple_test_example.go    # 简单测试示例
│   ├── tool_call_example.go      # 工具调用示例
│   ├── ui_example.go             # UI 示例
│   └── usage_example.go          # 使用示例
├── internal/                     # 内部包
│   ├── agent/                    # 代理相关
│   ├── analyzer/                 # 分析器
│   │   ├── config.go             # 配置
│   │   └── java_analyzer.go      # Java 分析器
│   ├── debug/                    # 调试工具
│   │   ├── logger.go             # 日志器
│   │   ├── logger_test.go        # 日志器测试
│   │   └── profiler.go           # 性能分析器
│   ├── llm/                      # 大语言模型
│   │   ├── client.go             # 客户端
│   │   └── openai.go             # OpenAI 实现
│   ├── monitor/                  # 监控
│   ├── supervisor/               # 监督器
│   ├── tools/                    # 工具
│   │   ├── analyzer_tools.go     # 分析器工具
│   │   └── tail.go               # 文件读取工具
│   └── ui/                       # 用户界面
│       └── chat.go               # 聊天界面
├── pkg/                          # 公共包
│   ├── logparser/                # 日志解析器
│   │   └── java_parser.go        # Java 日志解析器
│   └── reporter/                 # 报告器
├── build.sh                      # 构建脚本
├── build-linux4x.sh              # Linux 4.x 构建脚本
├── config.yaml.example           # 配置示例
├── CROSS_COMPILE.md              # 交叉编译指南
├── CROSS_COMPILE_SUMMARY.md      # 交叉编译总结
├── DEBUG.md                      # 调试文档
├── DEBUG_SUMMARY.md              # 调试总结
├── debug.sh                      # 调试脚本
├── demo.sh                       # 演示脚本
├── go.mod                        # Go 模块文件
├── go.sum                        # Go 依赖锁定文件
├── IMPROVEMENTS.md               # 改进建议
├── main.go                       # 主程序入口
├── Makefile                      # 构建配置
├── PROJECT_SUMMARY.md            # 项目总结
├── PROJECT_STRUCTURE.md          # 项目结构说明（本文件）
├── README.md                     # 项目说明
├── test-config.yaml              # 测试配置
├── USAGE.md                      # 使用说明
└── verify-build.sh               # 构建验证脚本
```

## 文件说明

### 核心文件
- `main.go`: 程序入口点
- `go.mod` / `go.sum`: Go 模块和依赖管理
- `Makefile`: 构建和开发工具配置

### 构建脚本
- `build.sh`: 完整的多平台构建脚本
- `build-linux4x.sh`: Linux 4.x 内核兼容性构建脚本
- `verify-build.sh`: 构建产物验证脚本

### 配置和文档
- `config.yaml.example`: 配置文件示例
- `README.md`: 项目主要说明文档
- `USAGE.md`: 详细使用说明
- `CROSS_COMPILE.md`: 交叉编译指南
- `DEBUG.md`: 调试相关文档

### 示例文件
- `examples/`: 包含各种使用示例
  - 聊天历史示例
  - 对话历史示例
  - 工具调用示例
  - UI 示例
  - 日志文件示例

### 内部包
- `internal/analyzer/`: 核心分析器逻辑
- `internal/llm/`: 大语言模型集成
- `internal/tools/`: 工具函数
- `internal/ui/`: 用户界面
- `internal/debug/`: 调试工具

### 公共包
- `pkg/logparser/`: 日志解析器
- `pkg/reporter/`: 报告生成器

## 清理说明

已清理的文件：
- 调试二进制文件 (`__debug_bin*`, `java-analyzer*`)
- 重复的测试文件
- 构建报告文件
- 空的目录

## 开发建议

1. **添加新功能**: 在 `internal/` 目录下创建相应的包
2. **添加示例**: 在 `examples/` 目录下创建示例文件
3. **添加测试**: 使用 `*_test.go` 命名约定
4. **构建**: 使用 `make build-all` 或 `./build.sh`
5. **验证**: 使用 `make verify` 验证构建产物

## 注意事项

- 所有示例文件都在 `examples/` 目录中
- 构建产物会被忽略（在 `.gitignore` 中）
- 配置文件示例使用 `.example` 后缀
- 调试相关文件使用 `debug` 前缀

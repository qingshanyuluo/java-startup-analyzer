# Java Startup Analyzer

一个基于Eino框架的Java程序启动失败分析工具，使用LLM来智能分析Java应用程序的启动日志，识别启动失败的原因并提供解决建议。

## 功能特性

- 🔍 **智能分析**: 使用LLM技术智能分析Java启动日志
- 🤖 **多模型支持**: 支持OpenAI、Anthropic等多种LLM提供商
- 📊 **详细报告**: 生成详细的分析报告和解决建议
- 🎨 **交互式聊天**: 提供智能聊天界面，支持问答交互
- ⚡ **自动分析**: 启动后自动分析配置的日志文件
- 🔧 **解决方案**: 提供具体的修复步骤和建议
- 📁 **Git集成**: 可选集成Git仓库进行代码分析

## 安装

```bash
go build -o java-analyzer main.go
```

## 使用方法

### 交互式聊天模式

```bash
# 使用配置文件启动聊天模式
./java-analyzer chat --config config.yaml
```

在聊天模式中，工具会：
- 自动读取配置文件中的启动命令和日志路径
- 自动开始分析Java启动日志
- 分析完成后允许您进行交互式聊天
- 获得智能的诊断和修复建议
- 使用 Ctrl+C 退出

### 配置文件格式

创建 `config.yaml` 配置文件：

```yaml
# LLM配置
model: "openai"  # 或 "anthropic"
api_key: "your-api-key"
base_url: ""  # 可选，自定义API端点

# 必需配置
start_cmd: "java -jar myapp.jar"  # Java启动命令
log_path: "/path/to/application.log"  # 日志文件路径

# 可选配置
git_repo: "/path/to/git/repository"  # Git仓库路径（可选）
verbose: false  # 详细输出模式
```

## 快速开始

1. 创建配置文件 `config.yaml`，填入您的配置信息
2. 确保日志文件存在且可读
3. 运行命令：`./java-analyzer chat --config config.yaml`
4. 工具会自动分析日志并进入聊天模式

## 支持的LLM提供商

- OpenAI GPT
- 其他兼容OpenAI API的模型

## 许可证

MIT License

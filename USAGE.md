# Java Startup Analyzer 使用说明

## 简介

Java Startup Analyzer 是一个基于Eino框架的智能Java程序启动失败分析工具。它使用大语言模型(LLM)来智能分析Java应用程序的启动日志，识别启动失败的原因并提供专业的解决建议。

## 安装

### 从源码构建

```bash
# 克隆项目
git clone <repository-url>
cd java-startup-analyzer

# 构建项目
./build.sh

# 或者手动构建
go build -o java-analyzer main.go
```

### 使用预编译版本

下载对应平台的预编译二进制文件，解压后即可使用。

## 基本使用

### 分析日志文件

```bash
# 分析Java启动日志文件
./java-analyzer analyze -f /path/to/java.log

# 分析示例日志
./java-analyzer analyze -f examples/sample-java-error.log
```

### 从标准输入读取

```bash
# 从标准输入读取日志
cat java.log | ./java-analyzer analyze

# 或者使用管道
tail -f application.log | ./java-analyzer analyze
```

### 输出格式

支持三种输出格式：

```bash
# 文本格式（默认）
./java-analyzer analyze -f java.log

# JSON格式
./java-analyzer analyze -f java.log --format json

# Markdown格式
./java-analyzer analyze -f java.log --format markdown
```

### 输出到文件

```bash
# 输出到文件
./java-analyzer analyze -f java.log -o analysis-report.txt

# JSON格式输出到文件
./java-analyzer analyze -f java.log --format json -o report.json
```

## 配置

### 命令行参数

```bash
# 设置LLM模型
./java-analyzer analyze -f java.log --model openai

# 设置API密钥
./java-analyzer analyze -f java.log --api-key YOUR_API_KEY

# 设置API基础URL
./java-analyzer analyze -f java.log --base-url https://api.openai.com/v1

# 详细输出模式
./java-analyzer analyze -f java.log --verbose
```

### 配置文件

创建配置文件 `~/.java-analyzer.yaml`：

```yaml
# LLM模型配置
model: "openai"
api_key: "your-api-key-here"
base_url: ""

# 输出配置
verbose: false

# 分析配置
analysis:
  max_log_size: 10485760  # 10MB
  timeout: 300            # 5分钟
  confidence_threshold: 0.5
```

### 环境变量

```bash
# 设置API密钥
export JAVA_ANALYZER_API_KEY="your-api-key-here"

# 设置模型类型
export JAVA_ANALYZER_MODEL="openai"
```

## 支持的错误类型

工具能够识别和分析以下常见的Java启动问题：

### 1. 类路径问题
- `ClassNotFoundException`
- `NoClassDefFoundError`
- 依赖缺失问题

### 2. 内存问题
- `OutOfMemoryError`
- 堆内存不足
- 内存泄漏

### 3. 网络和端口问题
- 端口占用
- 网络连接问题
- 防火墙配置

### 4. 配置文件问题
- 配置文件缺失
- 配置格式错误
- 环境变量问题

### 5. 权限问题
- 文件权限不足
- 目录访问权限
- 系统权限问题

## 输出说明

### 分析结果字段

- **状态**: success/failure/warning
- **错误类型**: 具体的错误类型
- **错误消息**: 主要的错误消息
- **根本原因**: 启动失败的根本原因分析
- **摘要**: 简要的问题描述
- **置信度**: 分析结果的置信度 (0-1)
- **解决建议**: 具体的修复建议列表

### 输出格式示例

#### 文本格式
```
=== Java启动失败分析报告 ===

📊 基本信息:
  分析时间: 2024-01-15 14:22:10
  日志大小: 5846 字符
  分析耗时: 1.25ms
  置信度: 85.0%

🔍 分析结果:
  状态: ❌ failure
  错误类型: OutOfMemoryError
  错误消息: Java heap space
  根本原因: 应用程序内存需求超过可用堆内存

💡 解决建议:
  1. 增加JVM堆内存大小 (-Xmx参数)
  2. 检查应用程序是否存在内存泄漏
  3. 优化应用程序的内存使用
```

#### JSON格式
```json
{
  "timestamp": "2024-01-15T14:22:10Z",
  "log_size": 5846,
  "analysis_time": 1250000,
  "status": "failure",
  "error_type": "OutOfMemoryError",
  "error_message": "Java heap space",
  "root_cause": "应用程序内存需求超过可用堆内存",
  "suggestions": [
    "增加JVM堆内存大小 (-Xmx参数)",
    "检查应用程序是否存在内存泄漏",
    "优化应用程序的内存使用"
  ],
  "confidence": 0.85
}
```

## 高级用法

### 批量分析

```bash
# 分析多个日志文件
for log in logs/*.log; do
  echo "分析文件: $log"
  ./java-analyzer analyze -f "$log" -o "reports/$(basename "$log").report"
done
```

### 集成到CI/CD

```bash
# 在CI/CD管道中使用
if ! ./java-analyzer analyze -f build.log --format json | jq -e '.status == "success"'; then
  echo "构建失败，请检查日志"
  exit 1
fi
```

### 监控模式

```bash
# 实时监控日志
tail -f application.log | ./java-analyzer analyze --format json | jq '.status'
```

## 故障排除

### 常见问题

1. **API密钥错误**
   ```
   错误: 创建LLM客户端失败
   解决: 检查API密钥是否正确设置
   ```

2. **网络连接问题**
   ```
   错误: 无法连接到LLM服务
   解决: 检查网络连接和API基础URL
   ```

3. **日志文件过大**
   ```
   错误: 日志文件超过最大限制
   解决: 使用tail命令截取最近的日志
   ```

### 调试模式

```bash
# 启用详细输出
./java-analyzer analyze -f java.log --verbose

# 查看帮助信息
./java-analyzer --help
./java-analyzer analyze --help
```

## 贡献

欢迎提交Issue和Pull Request来改进这个工具。

## 许可证

MIT License

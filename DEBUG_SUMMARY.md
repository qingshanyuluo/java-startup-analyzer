# Java Startup Analyzer 调试功能总结

## 🎯 调试功能概览

您的Java Startup Analyzer项目现在已经集成了完整的调试功能，包括：

### ✅ 已实现的调试功能

1. **调试日志系统** - 多级别日志记录
2. **性能分析器** - 内置pprof支持
3. **调试脚本** - 自动化调试工具
4. **VS Code调试配置** - IDE集成调试
5. **Makefile调试命令** - 便捷的调试命令
6. **环境变量支持** - 灵活的调试配置

## 🛠️ 调试方法

### 1. 基本调试

```bash
# 启用调试模式
DEBUG=true ./java-analyzer analyze -f error.log --verbose

# 启用性能分析
DEBUG_PROFILER=true ./java-analyzer analyze -f error.log

# 设置调试级别
DEBUG_LEVEL=debug ./java-analyzer analyze -f error.log
```

### 2. 使用调试脚本

```bash
# 查看所有调试命令
./debug.sh help

# 构建调试版本
./debug.sh build

# 运行测试
./debug.sh test

# 内存分析
./debug.sh memory

# CPU分析
./debug.sh cpu
```

### 3. 使用Makefile命令

```bash
# 构建调试版本
make debug-build

# 调试分析命令
make debug-analyze

# 调试聊天模式
make debug-chat

# 性能分析
make debug-profile

# 运行所有调试检查
make debug-all
```

### 4. IDE调试

在VS Code中：
1. 打开 `.vscode/launch.json`
2. 选择调试配置（如"Debug Analyze Command"）
3. 按F5开始调试

## 📊 调试功能详解

### 调试日志系统

- **多级别日志**: DEBUG, INFO, WARN, ERROR
- **环境变量控制**: DEBUG, DEBUG_LEVEL, VERBOSE
- **调用者信息**: 自动显示文件名和行号
- **时间戳**: 精确到毫秒的时间记录

### 性能分析器

- **内置pprof**: 自动启动性能分析服务器
- **Web界面**: 访问 http://localhost:6060 查看分析结果
- **多种分析**: CPU、内存、Goroutine、阻塞等
- **运行时统计**: 实时显示程序运行状态

### 调试脚本功能

- **构建调试版本**: 包含调试符号的构建
- **竞态检测**: 检测并发问题
- **代码质量检查**: 格式化和静态分析
- **性能分析**: 内存和CPU分析
- **聊天模式调试**: 交互式调试

## 🔧 调试配置

### 环境变量

```bash
# 基本调试
export DEBUG=true
export VERBOSE=true
export DEBUG_LEVEL=debug

# 性能分析
export DEBUG_PROFILER=true
export DEBUG_PROFILER_PORT=6060

# API配置
export JAVA_ANALYZER_API_KEY=your-api-key
```

### 配置文件

创建 `.java-analyzer.yaml` 进行调试配置：

```yaml
# 调试配置
model: "openai"
api_key: "your-api-key"
verbose: true

# 调试特定配置
debug:
  log_level: "debug"
  enable_tracing: true
  mock_llm: false
  timeout: 60
```

## 🚀 使用示例

### 调试分析命令

```bash
# 基本调试
DEBUG=true ./java-analyzer analyze -f examples/sample-java-error.log --verbose

# 使用调试脚本
./debug.sh build
./debug.sh test

# 使用Makefile
make debug-analyze
```

### 调试聊天模式

```bash
# 设置API密钥
export JAVA_ANALYZER_API_KEY=your-api-key

# 启动调试聊天模式
DEBUG=true ./java-analyzer chat --verbose

# 或使用脚本
./debug.sh chat
```

### 性能分析

```bash
# 启动性能分析
DEBUG_PROFILER=true ./java-analyzer analyze -f error.log

# 访问分析界面
open http://localhost:6060

# 使用Makefile
make debug-profile
```

## 📈 调试最佳实践

1. **分层调试**: 从简单到复杂，逐步调试
2. **日志记录**: 在关键位置添加适当的日志
3. **性能监控**: 定期检查性能指标
4. **错误处理**: 实现完善的错误处理机制
5. **测试覆盖**: 为每个组件编写测试

## 🔍 常见问题排查

### 聊天模式问题

```bash
# 检查终端兼容性
echo $TERM

# 使用最小化配置测试
./java-analyzer chat --api-key test --verbose

# 检查Bubble Tea依赖
go mod verify
```

### 性能问题

```bash
# 内存分析
make debug-memory

# CPU分析
make debug-cpu

# 执行跟踪
make debug-trace
```

### 配置问题

```bash
# 验证配置文件
./java-analyzer --config .java-analyzer.yaml --help

# 检查环境变量
env | grep JAVA_ANALYZER
```

## 📚 相关文件

- `DEBUG.md` - 详细调试指南
- `debug.sh` - 调试脚本
- `.vscode/launch.json` - VS Code调试配置
- `internal/debug/` - 调试功能实现
- `Makefile` - 调试相关命令

## 🎉 总结

您的Java Startup Analyzer项目现在拥有了完整的调试功能，包括：

- ✅ 多级别调试日志
- ✅ 性能分析器
- ✅ 自动化调试脚本
- ✅ IDE集成调试
- ✅ 环境变量配置
- ✅ 测试支持

这些功能将大大提高您的开发和调试效率，帮助您快速定位和解决问题！

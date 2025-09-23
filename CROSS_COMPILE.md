# 交叉编译指南

本文档介绍如何为 Java Startup Analyzer 进行交叉编译，特别是针对 Linux 4.x 内核的 x86 架构。

## 快速开始

### 使用 Makefile

```bash
# 构建所有平台版本
make build-all

# 构建 Linux 4.x 内核兼容版本
make build-linux4x

# 构建静态链接版本
make build-static

# 验证构建产物
make verify

# 完整发布准备
make release
```

### 使用构建脚本

```bash
# 构建所有版本
./build.sh

# 仅构建 Linux 4.x 兼容版本
./build-linux4x.sh

# 验证构建产物
./verify-build.sh
```

## 构建产物说明

### Linux 版本

| 文件名 | 架构 | 说明 | 适用场景 |
|--------|------|------|----------|
| `java-analyzer-linux-amd64` | x86_64 | 标准版本 | 现代 Linux 系统 |
| `java-analyzer-linux-amd64-kernel4x` | x86_64 | 内核 4.x 兼容 | Linux 4.x 内核系统 |
| `java-analyzer-linux-amd64-static` | x86_64 | 静态链接 | 旧版 Linux 系统 |

**注意**: 由于依赖库 `github.com/bytedance/sonic` 不支持 32 位架构，因此不提供 32 位 (x86) 版本的构建。

### 其他平台

| 文件名 | 平台 | 架构 |
|--------|------|------|
| `java-analyzer-darwin-amd64` | macOS | x86_64 |
| `java-analyzer-darwin-arm64` | macOS | ARM64 |
| `java-analyzer-windows-amd64.exe` | Windows | x86_64 |
| `java-analyzer-windows-386.exe` | Windows | x86 |

## 编译选项说明

### 标准编译

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-linux-amd64 main.go
```

- `CGO_ENABLED=0`: 禁用 CGO，确保纯 Go 编译
- `GOOS=linux`: 目标操作系统
- `GOARCH=amd64`: 目标架构
- `-ldflags="-w -s"`: 去除调试信息和符号表，减小文件大小

### Linux 4.x 内核兼容编译

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-w -s -extldflags '-static'" -o java-analyzer-linux-amd64-kernel4x main.go
```

- `-tags netgo`: 使用纯 Go 网络实现，避免 cgo 依赖
- `-extldflags '-static'`: 静态链接外部库

### 静态链接编译

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s -extldflags '-static'" -o java-analyzer-linux-amd64-static main.go
```

- `-a`: 强制重新构建所有包
- `-installsuffix cgo`: 使用不同的安装后缀，避免与 cgo 版本冲突

## 部署指南

### 1. 选择合适的版本

根据目标系统选择对应的二进制文件：

- **现代 Linux 系统 (内核 5.x+)**: 使用标准版本
- **Linux 4.x 内核系统**: 使用 `-kernel4x` 版本
- **旧版 Linux 系统**: 使用 `-static` 版本

### 2. 部署到目标系统

```bash
# 复制二进制文件
scp java-analyzer-linux-amd64-kernel4x user@target:/usr/local/bin/java-analyzer

# 设置执行权限
ssh user@target "chmod +x /usr/local/bin/java-analyzer"

# 测试运行
ssh user@target "java-analyzer --help"
```

### 3. 验证部署

```bash
# 检查文件类型
file /usr/local/bin/java-analyzer

# 检查依赖
ldd /usr/local/bin/java-analyzer

# 测试功能
java-analyzer analyze -f /path/to/logfile.log
```

## 故障排除

### 常见问题

1. **"cannot find package" 错误**
   - 确保运行了 `go mod tidy` 下载依赖

2. **"exec format error" 错误**
   - 检查目标架构是否正确
   - 确保使用了正确的 `GOARCH` 参数

3. **"no such file or directory" 错误**
   - 检查目标系统是否有必要的系统库
   - 尝试使用静态链接版本

4. **网络相关错误**
   - 使用 `-tags netgo` 确保使用纯 Go 网络实现

### 验证构建产物

使用验证脚本检查构建产物：

```bash
./verify-build.sh
```

该脚本会检查：
- 文件是否存在
- 文件类型和架构
- 依赖关系
- 执行权限

## 性能优化

### 减小文件大小

- 使用 `-ldflags="-w -s"` 去除调试信息
- 使用 `-ldflags="-s"` 去除符号表
- 使用 `upx` 工具进一步压缩（可选）

### 提高兼容性

- 禁用 CGO (`CGO_ENABLED=0`)
- 使用静态链接 (`-extldflags '-static'`)
- 使用纯 Go 网络实现 (`-tags netgo`)

## 开发建议

1. **测试多个版本**: 在构建后测试不同版本的功能
2. **验证兼容性**: 在目标系统上验证二进制文件
3. **文档化**: 记录特定版本的构建参数和适用场景
4. **自动化**: 使用 CI/CD 自动构建和测试

## 相关资源

- [Go 交叉编译文档](https://golang.org/doc/install/source#environment)
- [Go 构建约束](https://golang.org/pkg/go/build/#hdr-Build_Constraints)
- [静态链接指南](https://github.com/golang/go/wiki/cgo#static-linking)

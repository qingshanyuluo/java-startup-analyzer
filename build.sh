#!/bin/bash

# Java Startup Analyzer 构建脚本

set -e

echo "🔨 开始构建 Java Startup Analyzer..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go 1.18或更高版本"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.18"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ 错误: Go版本过低，需要1.18或更高版本，当前版本: $GO_VERSION"
    exit 1
fi

echo "✅ Go版本检查通过: $GO_VERSION"

# 清理之前的构建
echo "🧹 清理之前的构建..."
rm -f java-analyzer
rm -rf dist/

# 下载依赖
echo "📦 下载依赖..."
go mod tidy

# 运行测试
echo "🧪 运行测试..."
go test ./...

# 构建二进制文件
echo "🔨 构建二进制文件..."

# Linux 版本
echo "构建 Linux AMD64 (x86_64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-linux-amd64 main.go

echo "⚠️  跳过 Linux 386 构建 (依赖库不支持 32 位架构)"

# Linux 4.x 内核兼容版本
echo "构建 Linux AMD64 (兼容内核 4.x)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-w -s -extldflags '-static'" -o java-analyzer-linux-amd64-kernel4x main.go

echo "⚠️  跳过 Linux 386 内核 4.x 构建 (依赖库不支持 32 位架构)"

# 静态链接版本
echo "构建 Linux AMD64 静态版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s -extldflags '-static'" -o java-analyzer-linux-amd64-static main.go

echo "⚠️  跳过 Linux 386 静态版本构建 (依赖库不支持 32 位架构)"

# macOS 版本
echo "构建 Darwin AMD64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-darwin-amd64 main.go

echo "构建 Darwin ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o java-analyzer-darwin-arm64 main.go

# Windows 版本
echo "构建 Windows AMD64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-windows-amd64.exe main.go

echo "构建 Windows 386..."
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags="-w -s" -o java-analyzer-windows-386.exe main.go

# 创建本地版本
echo "构建本地版本..."
go build -o java-analyzer main.go

echo "✅ 构建完成！"
echo ""
echo "📁 生成的文件:"
echo "  - java-analyzer (本地版本)"
echo ""
echo "  Linux 版本:"
echo "  - java-analyzer-linux-amd64 (x86_64)"
echo "  - java-analyzer-linux-amd64-kernel4x (兼容内核 4.x)"
echo "  - java-analyzer-linux-amd64-static (静态链接)"
echo ""
echo "  macOS 版本:"
echo "  - java-analyzer-darwin-amd64"
echo "  - java-analyzer-darwin-arm64"
echo ""
echo "  Windows 版本:"
echo "  - java-analyzer-windows-amd64.exe"
echo "  - java-analyzer-windows-386.exe"
echo ""
echo "🚀 使用方法:"
echo "  ./java-analyzer analyze -f examples/sample-java-error.log"
echo ""
echo "📖 更多帮助:"
echo "  ./java-analyzer --help"
echo ""
echo "🔧 Linux 4.x 内核兼容性说明:"
echo "  - 使用 -kernel4x 后缀的版本专门为 Linux 4.x 内核优化"
echo "  - 使用 -static 后缀的版本为静态链接，适用于旧版 Linux 系统"
echo "  - 推荐在旧版 Linux 系统上使用静态链接版本"

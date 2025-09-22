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
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-linux-amd64 main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-darwin-amd64 main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o java-analyzer-darwin-arm64 main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o java-analyzer-windows-amd64.exe main.go

# 创建本地版本
go build -o java-analyzer main.go

echo "✅ 构建完成！"
echo ""
echo "📁 生成的文件:"
echo "  - java-analyzer (本地版本)"
echo "  - java-analyzer-linux-amd64"
echo "  - java-analyzer-darwin-amd64"
echo "  - java-analyzer-darwin-arm64"
echo "  - java-analyzer-windows-amd64.exe"
echo ""
echo "🚀 使用方法:"
echo "  ./java-analyzer analyze -f examples/sample-java-error.log"
echo ""
echo "📖 更多帮助:"
echo "  ./java-analyzer --help"

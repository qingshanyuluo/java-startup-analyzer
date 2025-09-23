#!/bin/bash

# Java Startup Analyzer Linux 4.x 内核兼容性构建脚本

set -e

echo "🐧 开始构建 Linux 4.x 内核兼容版本..."

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
rm -f java-analyzer-linux-*-kernel4x
rm -f java-analyzer-linux-*-static

# 下载依赖
echo "📦 下载依赖..."
go mod tidy

# 运行测试
echo "🧪 运行测试..."
go test ./...

# 构建 Linux 4.x 内核兼容版本
echo "🔨 构建 Linux 4.x 内核兼容版本..."

# AMD64 版本
echo "构建 Linux AMD64 (兼容内核 4.x)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags netgo \
    -ldflags="-w -s -extldflags '-static'" \
    -o java-analyzer-linux-amd64-kernel4x main.go

echo "⚠️  跳过 Linux 386 构建 (依赖库不支持 32 位架构)"

# 构建静态链接版本
echo "🔨 构建静态链接版本..."

# AMD64 静态版本
echo "构建 Linux AMD64 静态版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -o java-analyzer-linux-amd64-static main.go

echo "⚠️  跳过 Linux 386 静态版本构建 (依赖库不支持 32 位架构)"

echo "✅ Linux 4.x 内核兼容版本构建完成！"
echo ""
echo "📁 生成的文件:"
echo "  - java-analyzer-linux-amd64-kernel4x (AMD64, 兼容内核 4.x)"
echo "  - java-analyzer-linux-amd64-static (AMD64, 静态链接)"
echo ""

# 验证构建产物
echo "🔍 验证构建产物..."
for binary in java-analyzer-linux-*-kernel4x java-analyzer-linux-*-static; do
    if [ -f "$binary" ]; then
        echo "📄 $binary:"
        if command -v file >/dev/null 2>&1; then
            file "$binary"
        fi
        
        # 检查文件大小
        file_size=$(ls -lh "$binary" | awk '{print $5}')
        echo "📏 文件大小: $file_size"
        
        # 检查依赖
        if command -v ldd >/dev/null 2>&1; then
            echo "📋 依赖检查:"
            if ldd "$binary" 2>/dev/null | grep -q "not a dynamic executable"; then
                echo "✅ 静态链接"
            else
                echo "📋 动态链接依赖:"
                ldd "$binary" 2>/dev/null | head -3
            fi
        fi
        echo ""
    fi
done

echo "🎯 使用建议:"
echo "  - 对于 Linux 4.x 内核系统，使用 -kernel4x 版本"
echo "  - 对于旧版 Linux 系统，使用 -static 版本"
echo "  - 对于现代 Linux 系统，可以使用标准版本"
echo ""
echo "🚀 部署命令:"
echo "  # 复制到目标系统"
echo "  scp java-analyzer-linux-amd64-kernel4x user@target:/usr/local/bin/java-analyzer"
echo "  # 设置执行权限"
echo "  chmod +x /usr/local/bin/java-analyzer"
echo "  # 测试运行"
echo "  java-analyzer --help"
echo ""
echo "🔧 内核兼容性说明:"
echo "  - kernel4x 版本使用 netgo 标签，避免 cgo 依赖"
echo "  - static 版本完全静态链接，适用于 glibc 版本较旧的系统"
echo "  - 两个版本都禁用了 CGO，确保最大兼容性"

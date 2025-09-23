#!/bin/bash

# Java Startup Analyzer 交叉编译验证脚本

set -e

echo "🔍 开始验证交叉编译构建产物..."

# 检查构建目录
BUILD_DIR="build"
if [ ! -d "$BUILD_DIR" ]; then
    echo "❌ 错误: 构建目录 $BUILD_DIR 不存在"
    echo "请先运行 'make build-all' 或 './build.sh' 进行构建"
    exit 1
fi

# 验证函数
verify_binary() {
    local binary_path="$1"
    local expected_arch="$2"
    local expected_os="$3"
    
    if [ ! -f "$binary_path" ]; then
        echo "❌ 文件不存在: $binary_path"
        return 1
    fi
    
    # 检查文件类型
    if command -v file >/dev/null 2>&1; then
        local file_info=$(file "$binary_path")
        echo "📄 $binary_path: $file_info"
        
        # 验证架构
        if [[ "$file_info" == *"$expected_arch"* ]] || [[ "$file_info" == *"x86_64"* ]] && [[ "$expected_arch" == "x86-64" ]]; then
            echo "✅ 架构验证通过: $expected_arch"
        elif [[ "$file_info" == *"$expected_arch"* ]]; then
            echo "✅ 架构验证通过: $expected_arch"
        else
            echo "❌ 架构验证失败: 期望 $expected_arch，实际: $file_info"
            return 1
        fi
        
        # 验证操作系统
        if [[ "$file_info" == *"$expected_os"* ]] || [[ "$file_info" == *"SYSV"* ]] && [[ "$expected_os" == "Linux" ]]; then
            echo "✅ 操作系统验证通过: $expected_os"
        elif [[ "$file_info" == *"$expected_os"* ]]; then
            echo "✅ 操作系统验证通过: $expected_os"
        else
            echo "❌ 操作系统验证失败: 期望 $expected_os，实际: $file_info"
            return 1
        fi
    else
        echo "⚠️  file 命令不可用，跳过文件类型检查"
    fi
    
    # 检查文件大小
    local file_size=$(ls -lh "$binary_path" | awk '{print $5}')
    echo "📏 文件大小: $file_size"
    
    # 检查是否可执行
    if [ -x "$binary_path" ]; then
        echo "✅ 文件可执行"
    else
        echo "❌ 文件不可执行"
        return 1
    fi
    
    echo ""
    return 0
}

# 验证所有构建产物
echo "🔍 验证 Linux 版本..."
verify_binary "$BUILD_DIR/java-analyzer-linux-amd64" "x86-64" "Linux"

echo "🔍 验证 Linux 4.x 内核兼容版本..."
verify_binary "$BUILD_DIR/java-analyzer-linux-amd64-kernel4x" "x86-64" "Linux"

echo "🔍 验证静态链接版本..."
verify_binary "$BUILD_DIR/java-analyzer-linux-amd64-static" "x86-64" "Linux"

echo "🔍 验证 macOS 版本..."
verify_binary "$BUILD_DIR/java-analyzer-darwin-amd64" "x86-64" "Mach-O"
verify_binary "$BUILD_DIR/java-analyzer-darwin-arm64" "arm64" "Mach-O"

echo "🔍 验证 Windows 版本..."
verify_binary "$BUILD_DIR/java-analyzer-windows-amd64.exe" "x86-64" "PE32+"

# 检查静态链接
echo "🔍 检查静态链接..."
if command -v ldd >/dev/null 2>&1; then
    echo "检查 Linux 版本依赖..."
    for binary in "$BUILD_DIR"/java-analyzer-linux-*; do
        if [[ "$binary" == *".exe" ]] || [[ "$binary" == *"darwin"* ]]; then
            continue
        fi
        echo "📋 $binary 依赖:"
        if ldd "$binary" 2>/dev/null | grep -q "not a dynamic executable"; then
            echo "✅ 静态链接"
        else
            echo "📋 动态链接依赖:"
            ldd "$binary" 2>/dev/null | head -5
        fi
        echo ""
    done
else
    echo "⚠️  ldd 命令不可用，跳过依赖检查"
fi

# 测试本地版本
echo "🔍 测试本地版本..."
if [ -f "$BUILD_DIR/java-analyzer" ]; then
    echo "📋 本地版本信息:"
    "$BUILD_DIR/java-analyzer" --version 2>/dev/null || echo "版本信息获取失败"
    echo ""
else
    echo "⚠️  本地版本不存在"
fi

# 生成验证报告
echo "📊 生成验证报告..."
REPORT_FILE="build-verification-report.txt"
cat > "$REPORT_FILE" << EOF
Java Startup Analyzer 交叉编译验证报告
生成时间: $(date)
Go 版本: $(go version)

构建产物列表:
$(ls -la "$BUILD_DIR"/java-analyzer-* 2>/dev/null || echo "无构建产物")

文件类型检查:
$(for file in "$BUILD_DIR"/java-analyzer-*; do
    if [ -f "$file" ]; then
        echo "$file:"
        file "$file" 2>/dev/null || echo "  文件类型检查失败"
        echo ""
    fi
done)

依赖检查 (Linux 版本):
$(for binary in "$BUILD_DIR"/java-analyzer-linux-*; do
    if [ -f "$binary" ] && [[ "$binary" != *".exe" ]]; then
        echo "$binary:"
        if command -v ldd >/dev/null 2>&1; then
            ldd "$binary" 2>/dev/null | head -3 || echo "  依赖检查失败"
        else
            echo "  ldd 命令不可用"
        fi
        echo ""
    fi
done)
EOF

echo "✅ 验证完成！"
echo "📄 验证报告已保存到: $REPORT_FILE"
echo ""
echo "🎯 推荐使用:"
echo "  - Linux 4.x 内核: java-analyzer-linux-*-kernel4x"
echo "  - 旧版 Linux 系统: java-analyzer-linux-*-static"
echo "  - 现代 Linux 系统: java-analyzer-linux-*"
echo ""
echo "🚀 部署建议:"
echo "  1. 将对应的二进制文件复制到目标系统"
echo "  2. 设置执行权限: chmod +x java-analyzer-linux-*"
echo "  3. 测试运行: ./java-analyzer-linux-* --help"

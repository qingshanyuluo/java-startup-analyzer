#!/bin/bash

# Java Startup Analyzer 演示脚本

echo "🚀 Java Startup Analyzer 演示"
echo "================================"
echo ""

# 检查是否已构建
if [ ! -f "./java-analyzer" ]; then
    echo "📦 正在构建项目..."
    go build -o java-analyzer main.go
    if [ $? -ne 0 ]; then
        echo "❌ 构建失败"
        exit 1
    fi
    echo "✅ 构建完成"
    echo ""
fi

echo "📋 可用命令："
echo "1. 命令行模式 - 分析示例日志"
echo "2. 交互式聊天模式"
echo "3. 查看帮助信息"
echo "4. 退出"
echo ""

while true; do
    read -p "请选择 (1-4): " choice
    case $choice in
        1)
            echo ""
            echo "🔍 命令行模式演示..."
            echo "分析示例Java错误日志："
            echo ""
            ./java-analyzer analyze -f examples/sample-java-error.log --api-key "demo-key" --verbose
            echo ""
            echo "按任意键继续..."
            read -n 1
            ;;
        2)
            echo ""
            echo "💬 启动交互式聊天模式..."
            echo "注意：需要有效的API密钥才能使用"
            echo "使用 Ctrl+C 退出聊天模式"
            echo ""
            read -p "请输入API密钥 (或按回车跳过): " api_key
            if [ -n "$api_key" ]; then
                ./java-analyzer chat --api-key "$api_key"
            else
                echo "跳过聊天模式演示"
            fi
            ;;
        3)
            echo ""
            ./java-analyzer --help
            echo ""
            echo "按任意键继续..."
            read -n 1
            ;;
        4)
            echo "👋 再见！"
            exit 0
            ;;
        *)
            echo "❌ 无效选择，请输入 1-4"
            ;;
    esac
    echo ""
    echo "📋 可用命令："
    echo "1. 命令行模式 - 分析示例日志"
    echo "2. 交互式聊天模式"
    echo "3. 查看帮助信息"
    echo "4. 退出"
    echo ""
done

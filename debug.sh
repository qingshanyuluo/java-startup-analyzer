#!/bin/bash

# Java Startup Analyzer 调试工具

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查Go环境
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go未安装或不在PATH中"
        exit 1
    fi
    print_success "Go版本: $(go version)"
}

# 构建调试版本
build_debug() {
    print_info "构建调试版本..."
    go build -gcflags="all=-N -l" -o java-analyzer-debug main.go
    print_success "调试版本构建完成: java-analyzer-debug"
}

# 构建竞态检测版本
build_race() {
    print_info "构建竞态检测版本..."
    go build -race -o java-analyzer-race main.go
    print_success "竞态检测版本构建完成: java-analyzer-race"
}

# 运行测试
run_tests() {
    print_info "运行测试..."
    go test -v ./...
    print_success "测试完成"
}

# 运行基准测试
run_benchmarks() {
    print_info "运行基准测试..."
    go test -bench=. -benchmem ./internal/analyzer/
    print_success "基准测试完成"
}

# 检查代码质量
check_code() {
    print_info "检查代码质量..."
    
    # 格式化检查
    if ! go fmt ./...; then
        print_warning "代码格式需要调整"
    else
        print_success "代码格式正确"
    fi
    
    # 静态分析
    if command -v golangci-lint &> /dev/null; then
        print_info "运行静态分析..."
        golangci-lint run
        print_success "静态分析完成"
    else
        print_warning "golangci-lint未安装，跳过静态分析"
    fi
}

# 内存分析
profile_memory() {
    print_info "启动内存分析..."
    
    # 构建支持pprof的版本
    go build -o java-analyzer main.go
    
    # 启动程序
    print_info "启动程序进行内存分析..."
    ./java-analyzer analyze -f examples/sample-java-error.log &
    PID=$!
    
    # 等待程序启动
    sleep 2
    
    # 启动pprof
    print_info "启动pprof服务器 (http://localhost:8080)"
    go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap &
    PPROF_PID=$!
    
    print_info "按Ctrl+C停止分析"
    wait $PID
    
    # 清理
    kill $PPROF_PID 2>/dev/null || true
}

# CPU分析
profile_cpu() {
    print_info "启动CPU分析..."
    
    # 构建支持pprof的版本
    go build -o java-analyzer main.go
    
    # 启动程序
    print_info "启动程序进行CPU分析..."
    ./java-analyzer analyze -f examples/sample-java-error.log &
    PID=$!
    
    # 等待程序启动
    sleep 2
    
    # 启动pprof
    print_info "启动pprof服务器 (http://localhost:8080)"
    go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile &
    PPROF_PID=$!
    
    print_info "按Ctrl+C停止分析"
    wait $PID
    
    # 清理
    kill $PPROF_PID 2>/dev/null || true
}

# 调试聊天模式
debug_chat() {
    print_info "调试聊天模式..."
    
    if [ -z "$JAVA_ANALYZER_API_KEY" ]; then
        print_warning "请设置JAVA_ANALYZER_API_KEY环境变量"
        read -p "请输入API密钥: " api_key
        export JAVA_ANALYZER_API_KEY="$api_key"
    fi
    
    print_info "启动聊天模式调试..."
    ./java-analyzer chat --verbose --api-key "$JAVA_ANALYZER_API_KEY"
}

# 验证配置
validate_config() {
    print_info "验证配置文件..."
    
    if [ -f ".java-analyzer.yaml" ]; then
        print_success "找到配置文件: .java-analyzer.yaml"
        ./java-analyzer --config .java-analyzer.yaml --help
    else
        print_warning "未找到配置文件，使用默认配置"
        ./java-analyzer --help
    fi
}

# 清理调试文件
cleanup() {
    print_info "清理调试文件..."
    rm -f java-analyzer-debug java-analyzer-race
    print_success "清理完成"
}

# 显示帮助
show_help() {
    echo "Java Startup Analyzer 调试工具"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "可用命令:"
    echo "  build       构建调试版本"
    echo "  race        构建竞态检测版本"
    echo "  test        运行测试"
    echo "  bench       运行基准测试"
    echo "  check       检查代码质量"
    echo "  memory      内存分析"
    echo "  cpu         CPU分析"
    echo "  chat        调试聊天模式"
    echo "  config      验证配置"
    echo "  clean       清理调试文件"
    echo "  help        显示此帮助信息"
    echo ""
    echo "环境变量:"
    echo "  JAVA_ANALYZER_API_KEY    LLM API密钥"
    echo "  DEBUG                   启用调试模式"
    echo "  VERBOSE                 启用详细输出"
}

# 主函数
main() {
    check_go
    
    case "${1:-help}" in
        "build")
            build_debug
            ;;
        "race")
            build_race
            ;;
        "test")
            run_tests
            ;;
        "bench")
            run_benchmarks
            ;;
        "check")
            check_code
            ;;
        "memory")
            profile_memory
            ;;
        "cpu")
            profile_cpu
            ;;
        "chat")
            debug_chat
            ;;
        "config")
            validate_config
            ;;
        "clean")
            cleanup
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# 运行主函数
main "$@"

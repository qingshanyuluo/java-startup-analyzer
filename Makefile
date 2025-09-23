# Java Startup Analyzer Makefile

.PHONY: build clean test install help run-example

# 默认目标
.DEFAULT_GOAL := help

# 变量
BINARY_NAME=java-analyzer
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -w -s"

# 帮助信息
help: ## 显示帮助信息
	@echo "Java Startup Analyzer 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 构建
build: ## 构建二进制文件
	@echo "🔨 构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建多平台版本
build-all: ## 构建多平台版本
	@echo "🔨 构建多平台版本..."
	@mkdir -p $(BUILD_DIR)
	@echo "构建 Linux AMD64 (x86_64)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "⚠️  跳过 Linux 386 构建 (依赖库不支持 32 位架构)"
	@echo "构建 Darwin AMD64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	@echo "构建 Darwin ARM64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@echo "构建 Windows AMD64..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "⚠️  跳过 Windows 386 构建 (依赖库不支持 32 位架构)"
	@echo "✅ 多平台构建完成"

# 构建 Linux 4.x 内核兼容版本
build-linux4x: ## 构建 Linux 4.x 内核兼容版本
	@echo "🔨 构建 Linux 4.x 内核兼容版本..."
	@mkdir -p $(BUILD_DIR)
	@echo "构建 Linux AMD64 (兼容内核 4.x)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-w -s -extldflags '-static'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64-kernel4x main.go
	@echo "⚠️  跳过 Linux 386 构建 (依赖库不支持 32 位架构)"
	@echo "✅ Linux 4.x 内核兼容版本构建完成"

# 构建静态链接版本（适用于旧版 Linux）
build-static: ## 构建静态链接版本
	@echo "🔨 构建静态链接版本..."
	@mkdir -p $(BUILD_DIR)
	@echo "构建 Linux AMD64 静态版本..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s -extldflags '-static'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64-static main.go
	@echo "⚠️  跳过 Linux 386 静态版本构建 (依赖库不支持 32 位架构)"
	@echo "✅ 静态链接版本构建完成"

# 安装
install: build ## 安装到系统
	@echo "📦 安装 $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ 安装完成"

# 测试
test: ## 运行测试
	@echo "🧪 运行测试..."
	go test -v ./internal/... ./pkg/...

# 测试覆盖率
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试覆盖率..."
	go test -v -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告生成: coverage.html"

# 清理
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ 清理完成"

# 运行示例
run-example: build ## 运行示例分析
	@echo "🚀 运行示例分析..."
	$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log

# 运行聊天模式演示
run-chat: build ## 运行聊天模式演示
	@echo "💬 启动聊天模式..."
	@echo "注意：需要设置API密钥环境变量"
	@echo "export JAVA_ANALYZER_API_KEY=your-api-key"
	$(BUILD_DIR)/$(BINARY_NAME) chat --api-key $(JAVA_ANALYZER_API_KEY)

# 运行演示脚本
run-demo: build ## 运行交互式演示脚本
	@echo "🎮 启动演示脚本..."
	./demo.sh

# 运行内存错误示例
run-memory-example: build ## 运行内存错误示例
	@echo "🚀 运行内存错误示例..."
	$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/out-of-memory-error.log

# 运行JSON输出示例
run-json-example: build ## 运行JSON输出示例
	@echo "🚀 运行JSON输出示例..."
	$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log --format json

# 运行Markdown输出示例
run-markdown-example: build ## 运行Markdown输出示例
	@echo "🚀 运行Markdown输出示例..."
	$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log --format markdown

# 依赖管理
deps: ## 下载依赖
	@echo "📦 下载依赖..."
	go mod download
	go mod tidy

# 格式化代码
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	go fmt ./...

# 代码检查
lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过代码检查"; \
		echo "   安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 安全检查
security: ## 运行安全检查
	@echo "🔒 运行安全检查..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec 未安装，跳过安全检查"; \
		echo "   安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 开发模式
dev: ## 开发模式（自动重新构建）
	@echo "🔄 开发模式启动..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air 未安装，使用普通构建模式"; \
		echo "   安装命令: go install github.com/cosmtrek/air@latest"; \
		make build && $(BUILD_DIR)/$(BINARY_NAME) --help; \
	fi

# 发布准备
release: clean test build-all build-linux4x build-static ## 准备发布版本
	@echo "📦 准备发布版本..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME)-* $(BUILD_DIR)/release/
	@cp README.md USAGE.md $(BUILD_DIR)/release/
	@cp config.yaml.example $(BUILD_DIR)/release/ 2>/dev/null || true
	@echo "✅ 发布版本准备完成: $(BUILD_DIR)/release/"
	@echo "📁 包含的构建产物:"
	@ls -la $(BUILD_DIR)/release/$(BINARY_NAME)-*

# 验证构建产物
verify: build-all build-linux4x build-static ## 验证构建产物
	@echo "🔍 验证构建产物..."
	@./verify-build.sh

# 显示版本信息
version: ## 显示版本信息
	@echo "版本: $(VERSION)"
	@echo "Go版本: $(shell go version)"
	@echo "构建时间: $(shell date)"

# 调试相关命令
debug-build: ## 构建调试版本
	@echo "🔧 构建调试版本..."
	@mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug main.go
	@echo "✅ 调试版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-debug"

debug-race: ## 构建竞态检测版本
	@echo "🔧 构建竞态检测版本..."
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-race main.go
	@echo "✅ 竞态检测版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-race"

debug-test: ## 运行调试测试
	@echo "🧪 运行调试测试..."
	DEBUG=true go test -v ./...

debug-analyze: debug-build ## 调试分析命令
	@echo "🔍 调试分析命令..."
	DEBUG=true $(BUILD_DIR)/$(BINARY_NAME)-debug analyze -f examples/sample-java-error.log --verbose

debug-chat: debug-build ## 调试聊天模式
	@echo "💬 调试聊天模式..."
	@echo "注意：需要设置JAVA_ANALYZER_API_KEY环境变量"
	DEBUG=true $(BUILD_DIR)/$(BINARY_NAME)-debug chat --verbose --api-key $(JAVA_ANALYZER_API_KEY)

debug-profile: ## 启动性能分析
	@echo "📊 启动性能分析..."
	@echo "访问 http://localhost:6060 查看性能分析"
	DEBUG_PROFILER=true $(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log

debug-memory: ## 内存分析
	@echo "🧠 内存分析..."
	@echo "访问 http://localhost:8080 查看内存分析"
	@$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log &
	@sleep 2
	@go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

debug-cpu: ## CPU分析
	@echo "⚡ CPU分析..."
	@echo "访问 http://localhost:8080 查看CPU分析"
	@$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log &
	@sleep 2
	@go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile

debug-trace: ## 执行跟踪
	@echo "📈 执行跟踪..."
	@$(BUILD_DIR)/$(BINARY_NAME) analyze -f examples/sample-java-error.log 2> trace.out
	@go tool trace trace.out

debug-callgraph: ## 调用图分析
	@echo "🕸️ 调用图分析..."
	@if command -v go-callvis >/dev/null 2>&1; then \
		go-callvis -group pkg,type -focus github.com/user/java-startup-analyzer .; \
	else \
		echo "请先安装 go-callvis: go install github.com/ofthehead/go-callvis@latest"; \
	fi

debug-all: debug-build debug-race debug-test ## 运行所有调试检查
	@echo "🔍 运行所有调试检查..."

# 完整构建流程
all: clean deps fmt lint test build-all ## 完整构建流程
	@echo "🎉 完整构建流程完成！"

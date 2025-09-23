package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/user/java-startup-analyzer/internal/analyzer"
)

func main() {
	// 创建配置
	config := &analyzer.Config{
		Model:     "openai",
		ModelName: "gpt-3.5-turbo",
		APIKey:    "your-api-key-here", // 请替换为实际的API密钥
		BaseURL:   "",                  // 使用默认的OpenAI API URL
	}

	// 创建分析器
	javaAnalyzer, err := analyzer.NewJavaAnalyzer(config)
	if err != nil {
		log.Fatalf("创建分析器失败: %v", err)
	}

	// 创建一个测试日志文件
	testLogPath := "test-chat-history-fix.log"
	testLogContent := `
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.SpringApplication - Application startup failed
java.lang.OutOfMemoryError: Java heap space
	at java.util.Arrays.copyOf(Arrays.java:3210)
	at java.util.ArrayList.grow(ArrayList.java:267)
	at com.example.MyApplication.loadData(MyApplication.java:45)
	at com.example.MyApplication.main(MyApplication.java:23)
`

	// 创建测试日志文件
	err = os.WriteFile(testLogPath, []byte(testLogContent), 0644)
	if err != nil {
		log.Fatalf("创建测试日志文件失败: %v", err)
	}
	defer os.Remove(testLogPath) // 清理测试文件

	fmt.Println("=== 测试聊天历史修复 ===")
	fmt.Printf("测试日志文件: %s\n", testLogPath)
	fmt.Println("现在应该能正确处理用户输入和历史消息了")
	fmt.Println(strings.Repeat("=", 50))

	ctx := context.Background()

	// 测试1: 初次分析日志文件
	fmt.Println("\n=== 测试1: 初次分析日志文件 ===")
	input1 := map[string]any{
		"log_path": testLogPath,
	}

	response1, err := javaAnalyzer.Chat(ctx, input1)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		fmt.Println("初次分析结果:")
		fmt.Println(response1.Content[:200] + "...")
	}

	// 测试2: 继续聊天 - 用户输入
	fmt.Println("\n=== 测试2: 继续聊天 - 用户输入 ===")
	input2 := map[string]any{
		"input": "这个错误的具体原因是什么？",
	}

	response2, err := javaAnalyzer.Chat(ctx, input2)
	if err != nil {
		log.Printf("继续聊天失败: %v", err)
	} else {
		fmt.Println("继续聊天结果:")
		fmt.Println(response2.Content[:200] + "...")
	}

	// 测试3: 再次继续聊天
	fmt.Println("\n=== 测试3: 再次继续聊天 ===")
	input3 := map[string]any{
		"input": "如何解决这个问题？",
	}

	response3, err := javaAnalyzer.Chat(ctx, input3)
	if err != nil {
		log.Printf("再次聊天失败: %v", err)
	} else {
		fmt.Println("再次聊天结果:")
		fmt.Println(response3.Content[:200] + "...")
	}

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("如果看到大模型能够正确响应后续的用户输入，说明修复成功！")
}

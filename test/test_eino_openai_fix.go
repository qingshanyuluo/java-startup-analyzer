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
	testLogPath := "test-eino-openai-fix.log"
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

	fmt.Println("=== 测试 Eino OpenAI 修复 ===")
	fmt.Printf("测试日志文件: %s\n", testLogPath)
	fmt.Println("现在使用 Eino 官方的 OpenAI 实现，应该能正确调用工具了")
	fmt.Println(strings.Repeat("=", 50))

	// 测试：使用文件路径，让大模型自己调用工具
	input := map[string]any{
		"log_path": testLogPath,
	}

	ctx := context.Background()
	response, err := javaAnalyzer.Chat(ctx, input)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		fmt.Println("分析结果:")
		fmt.Println(response.Content)
	}

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("如果看到大模型真正调用了 tail 工具，说明修复成功！")
}

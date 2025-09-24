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
		log.Fatalf("创建Java分析器失败: %v", err)
	}

	// 创建测试日志文件
	testLogFiles := []struct {
		name        string
		description string
		filePath    string
		logContent  string
	}{
		{
			name:        "内存不足错误",
			description: "测试OutOfMemoryError分析",
			filePath:    "test-oom-error.log",
			logContent: `
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.SpringApplication - Application startup failed
java.lang.OutOfMemoryError: Java heap space
	at java.util.Arrays.copyOf(Arrays.java:3210)
	at java.util.ArrayList.grow(ArrayList.java:267)
	at com.example.MyApplication.loadData(MyApplication.java:45)
`,
		},
		{
			name:        "类未找到错误",
			description: "测试ClassNotFoundException分析",
			filePath:    "test-class-not-found.log",
			logContent: `
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.SpringApplication - Application startup failed
java.lang.ClassNotFoundException: com.example.MissingClass
	at java.net.URLClassLoader.findClass(URLClassLoader.java:382)
	at java.lang.ClassLoader.loadClass(ClassLoader.java:424)
	at org.springframework.boot.loader.LaunchedURLClassLoader.loadClass(LaunchedURLClassLoader.java:151)
`,
		},
		{
			name:        "端口被占用",
			description: "测试端口冲突分析",
			filePath:    "test-port-in-use.log",
			logContent: `
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.web.embedded.tomcat.TomcatWebServer - Tomcat initialized with port(s): 8080 (http)
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.SpringApplication - Application startup failed
java.net.BindException: Address already in use: bind
	at sun.nio.ch.Net.bind0(Native Method)
	at sun.nio.ch.Net.bind(Net.java:433)
	at sun.nio.ch.Net.bind(Net.java:425)
`,
		},
	}

	ctx := context.Background()

	// 测试每个场景
	for i, testCase := range testLogFiles {
		fmt.Printf("\n=== 测试案例 %d: %s ===\n", i+1, testCase.name)
		fmt.Printf("描述: %s\n", testCase.description)
		fmt.Printf("日志文件: %s\n", testCase.filePath)
		fmt.Println("日志内容预览:")
		fmt.Println(testCase.logContent[:200] + "...")
		fmt.Println("\n分析结果:")
		fmt.Println(strings.Repeat("=", 50))

		// 创建测试日志文件
		err := os.WriteFile(testCase.filePath, []byte(testCase.logContent), 0644)
		if err != nil {
			log.Printf("创建测试日志文件失败: %v", err)
			continue
		}
		defer os.Remove(testCase.filePath) // 清理测试文件

		// 创建输入 - 使用文件路径而不是日志内容
		input := map[string]any{
			"log_path": testCase.filePath,
		}

		// 执行流式分析
		stream, err := javaAnalyzer.ChatStream(ctx, input)
		if err != nil {
			log.Printf("分析失败: %v", err)
			continue
		}
		defer stream.Close()

		fmt.Println("分析结果:")
		for {
			chunk, err := stream.Recv()
			if err != nil {
				break
			}
			fmt.Print(chunk.Content)
		}
		fmt.Println("\n" + strings.Repeat("=", 50))
	}

	// 测试流式分析
	fmt.Println("\n=== 流式分析测试 ===")

	// 创建流式分析测试文件
	streamTestLog := `
2024-01-15 10:30:15.123 INFO [main] com.example.MyApplication - Starting MyApplication
2024-01-15 10:30:15.123 INFO [main] com.example.MyApplication - No active profile set, falling back to default profiles: default
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.SpringApplication - Application startup failed
java.lang.IllegalArgumentException: Invalid configuration
	at com.example.MyApplication.validateConfig(MyApplication.java:67)
	at com.example.MyApplication.main(MyApplication.java:23)
`

	streamTestFile := "test-stream-analysis.log"
	err = os.WriteFile(streamTestFile, []byte(streamTestLog), 0644)
	if err != nil {
		log.Printf("创建流式测试日志文件失败: %v", err)
		return
	}
	defer os.Remove(streamTestFile) // 清理测试文件

	streamInput := map[string]any{
		"log_path": streamTestFile,
	}

	stream, err := javaAnalyzer.ChatStream(ctx, streamInput)
	if err != nil {
		log.Printf("创建流式分析失败: %v", err)
		return
	}
	defer stream.Close()

	fmt.Println("流式分析结果:")
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println("\n流式分析完成")
}

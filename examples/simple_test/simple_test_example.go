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
	testLogPath := "test-java-error.log"
	testLogContent := `
2024-01-15 10:30:15.123 ERROR [main] org.springframework.boot.SpringApplication - Application startup failed
java.lang.OutOfMemoryError: Java heap space
	at java.util.Arrays.copyOf(Arrays.java:3210)
	at java.util.ArrayList.grow(ArrayList.java:267)
	at java.util.ArrayList.ensureExplicitCapacity(ArrayList.java:241)
	at java.util.ArrayList.ensureCapacityInternal(ArrayList.java:233)
	at java.util.ArrayList.add(ArrayList.java:464)
	at com.example.MyApplication.loadData(MyApplication.java:45)
	at com.example.MyApplication.main(MyApplication.java:23)
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.springframework.boot.loader.MainMethodRunner.run(MainMethodRunner.java:49)
	at org.springframework.boot.loader.Launcher.launch(Launcher.java:95)
	at org.springframework.boot.loader.Launcher.launch(Launcher.java:58)
	at org.springframework.boot.loader.JarLauncher.main(JarLauncher.java:88)

2024-01-15 10:30:15.124 ERROR [main] org.springframework.boot.SpringApplication - Application run failed
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

	fmt.Println("=== 测试新的Java分析器 ===")
	fmt.Printf("测试日志文件: %s\n", testLogPath)
	fmt.Println("日志内容预览:")
	fmt.Println(testLogContent[:200] + "...")
	fmt.Println("\n" + strings.Repeat("=", 50))

	// 测试1: 使用文件路径进行分析
	fmt.Println("\n=== 测试1: 使用文件路径分析 ===")
	input1 := map[string]any{
		"log_path": testLogPath,
	}

	ctx := context.Background()
	response1, err := javaAnalyzer.Chat(ctx, input1)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		fmt.Println("分析结果:")
		fmt.Println(response1.Content)
	}

	// 测试2: 使用旧的log_content参数（应该提示使用文件路径）
	fmt.Println("\n=== 测试2: 使用log_content参数（应该提示使用文件路径）===")
	input2 := map[string]any{
		"log_content": "一些日志内容",
	}

	response2, err := javaAnalyzer.Chat(ctx, input2)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		fmt.Println("响应:")
		fmt.Println(response2.Content)
	}

	// 测试3: 流式分析
	fmt.Println("\n=== 测试3: 流式分析 ===")
	input3 := map[string]any{
		"log_path": testLogPath,
	}

	stream, err := javaAnalyzer.ChatStream(ctx, input3)
	if err != nil {
		log.Printf("创建流式分析失败: %v", err)
	} else {
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

	fmt.Println("\n=== 测试完成 ===")
}

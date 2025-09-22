package logparser

import (
	"regexp"
	"strings"
	"time"
)

// JavaLogInfo Java日志信息
type JavaLogInfo struct {
	StartupTime string   // 启动时间
	JavaVersion string   // Java版本
	MainClass   string   // 主类
	Errors      []string // 错误信息
	StackTraces []string // 异常堆栈
}

// JavaLogParser Java日志解析器
type JavaLogParser struct {
	// 正则表达式模式
	startupTimePattern *regexp.Regexp
	javaVersionPattern *regexp.Regexp
	mainClassPattern   *regexp.Regexp
	errorPattern       *regexp.Regexp
	stackTracePattern  *regexp.Regexp
}

// NewJavaLogParser 创建新的Java日志解析器
func NewJavaLogParser() *JavaLogParser {
	return &JavaLogParser{
		startupTimePattern: regexp.MustCompile(`(?i)(started|starting|launching).*?(\d{4}-\d{2}-\d{2}|\d{2}:\d{2}:\d{2})`),
		javaVersionPattern: regexp.MustCompile(`(?i)java version "([^"]+)"`),
		mainClassPattern:   regexp.MustCompile(`(?i)main class[:\s]+([a-zA-Z_$][a-zA-Z0-9_$\.]*)`),
		errorPattern:       regexp.MustCompile(`(?i)(error|exception|failed|failure):\s*(.+)`),
		stackTracePattern:  regexp.MustCompile(`(?i)(\w+\.\w+Exception|\w+\.\w+Error).*?\n((?:\s+at\s+.*\n?)*)`),
	}
}

// Parse 解析Java日志
func (p *JavaLogParser) Parse(logContent string) (*JavaLogInfo, error) {
	info := &JavaLogInfo{
		Errors:      make([]string, 0),
		StackTraces: make([]string, 0),
	}

	lines := strings.Split(logContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析启动时间
		if info.StartupTime == "" {
			if matches := p.startupTimePattern.FindStringSubmatch(line); len(matches) > 2 {
				info.StartupTime = matches[2]
			}
		}

		// 解析Java版本
		if info.JavaVersion == "" {
			if matches := p.javaVersionPattern.FindStringSubmatch(line); len(matches) > 1 {
				info.JavaVersion = matches[1]
			}
		}

		// 解析主类
		if info.MainClass == "" {
			if matches := p.mainClassPattern.FindStringSubmatch(line); len(matches) > 1 {
				info.MainClass = matches[1]
			}
		}

		// 解析错误信息
		if matches := p.errorPattern.FindStringSubmatch(line); len(matches) > 2 {
			info.Errors = append(info.Errors, strings.TrimSpace(matches[2]))
		}
	}

	// 解析异常堆栈
	stackTraces := p.stackTracePattern.FindAllStringSubmatch(logContent, -1)
	for _, match := range stackTraces {
		if len(match) > 2 {
			stackTrace := strings.TrimSpace(match[0])
			info.StackTraces = append(info.StackTraces, stackTrace)
		}
	}

	// 如果没有找到启动时间，尝试从日志开头提取
	if info.StartupTime == "" {
		info.StartupTime = p.extractStartupTimeFromLog(logContent)
	}

	return info, nil
}

// extractStartupTimeFromLog 从日志中提取启动时间
func (p *JavaLogParser) extractStartupTimeFromLog(logContent string) string {
	lines := strings.Split(logContent, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		// 尝试匹配常见的时间格式
		timePatterns := []string{
			`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`,
			`\d{2}:\d{2}:\d{2}`,
			`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`,
		}

		for _, pattern := range timePatterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(firstLine); len(matches) > 0 {
				return matches[0]
			}
		}
	}

	return time.Now().Format("2006-01-02 15:04:05")
}

// ExtractErrors 提取错误信息
func (p *JavaLogParser) ExtractErrors(logContent string) []string {
	errors := make([]string, 0)
	matches := p.errorPattern.FindAllStringSubmatch(logContent, -1)

	for _, match := range matches {
		if len(match) > 2 {
			error := strings.TrimSpace(match[2])
			errors = append(errors, error)
		}
	}

	return errors
}

// ExtractStackTraces 提取异常堆栈
func (p *JavaLogParser) ExtractStackTraces(logContent string) []string {
	stackTraces := make([]string, 0)
	matches := p.stackTracePattern.FindAllStringSubmatch(logContent, -1)

	for _, match := range matches {
		if len(match) > 0 {
			stackTrace := strings.TrimSpace(match[0])
			stackTraces = append(stackTraces, stackTrace)
		}
	}

	return stackTraces
}

// IsStartupFailure 判断是否为启动失败
func (p *JavaLogParser) IsStartupFailure(logContent string) bool {
	failureIndicators := []string{
		"failed to start",
		"startup failed",
		"could not start",
		"unable to start",
		"startup error",
		"initialization failed",
		"bootstrap failed",
	}

	lowerContent := strings.ToLower(logContent)
	for _, indicator := range failureIndicators {
		if strings.Contains(lowerContent, indicator) {
			return true
		}
	}

	// 检查是否有异常堆栈
	stackTraces := p.ExtractStackTraces(logContent)
	return len(stackTraces) > 0
}

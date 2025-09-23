package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudwego/eino/schema"
	"github.com/user/java-startup-analyzer/internal/analyzer"
)

// Message 表示聊天中的一条消息
type Message struct {
	Content string
	Sender  string // "user" 或 "bot"
	Time    time.Time
	Type    string // "text", "analysis", "error"
}

// ChatModel 聊天界面的模型
type ChatModel struct {
	messages       []Message
	viewport       viewport.Model
	input          string
	cursor         int
	isProcessing   bool
	processingText string
	analyzer       *analyzer.JavaAnalyzer
	config         *analyzer.Config
	ctrlCPressed   bool // 跟踪是否已经按了一次Ctrl+C
	wasInterrupted bool // 跟踪是否被Ctrl+C打断
	ready          bool
	streamingMsg   string                                // 当前流式输出的消息内容
	typingFull     string                                // 打字机效果的完整文本
	typingPos      int                                   // 当前已输出的位置（按rune计）
	isFirst        bool                                  // 是否是第一次分析
	streamReader   *schema.StreamReader[*schema.Message] // 流式读取器
}

// AnalysisCompleteMsg 分析完成的消息
type AnalysisCompleteMsg struct {
	Error error
}

// StreamMsg 流式输出消息
type StreamMsg struct {
	Content string
	Done    bool
	Error   error
}

type processingTickMsg time.Time
type startProcessingMsg struct{}

// typingTickMsg 触发打字机效果的定时消息
type typingTickMsg time.Time

// StartTypingMsg 启动打字机效果，携带完整内容
type StartTypingMsg struct {
	Content string
	isFirst bool
}

// StartStreamMsg 启动真正的流式处理
type StartStreamMsg struct {
	StreamReader *schema.StreamReader[*schema.Message]
	isFirst      bool
}

type analysisDoneMsg struct{}

// NewChatModel 创建新的聊天模型
func NewChatModel(config *analyzer.Config) (*ChatModel, error) {
	javaAnalyzer, err := analyzer.NewJavaAnalyzer(config)
	if err != nil {
		return nil, err
	}

	return &ChatModel{
		messages: []Message{
			{
				Content: "🤖 欢迎使用Java启动分析器！\n\n正在分析您的Java启动日志作为背景信息...",
				Sender:  "bot",
				Time:    time.Now(),
				Type:    "text",
			},
		},
		input:    "",
		analyzer: javaAnalyzer,
		config:   config,
	}, nil
}

func (m ChatModel) Init() tea.Cmd {
	// 启动时自动开始分析日志
	return tea.Sequence(func() tea.Msg { return startProcessingMsg{} }, m.autoAnalyze())
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-8)
			m.viewport.SetContent(m.renderMessages())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 8
		}

	case startProcessingMsg:
		m.isProcessing = true
		m.processingText = "正在分析Java日志..."
		return m, m.updateProcessingText()

	case StartTypingMsg:
		// 初始化打字机效果
		m.streamingMsg = ""
		m.typingFull = msg.Content
		m.typingPos = 0
		m.isFirst = msg.isFirst
		// 立即开始第一次输出
		return m, tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg { return typingTickMsg(t) })

	case StartStreamMsg:
		// 启动真正的流式处理
		m.streamingMsg = ""
		m.isFirst = msg.isFirst
		m.streamReader = msg.StreamReader
		// 启动流式读取
		return m, m.startStreaming(msg.StreamReader)

	case tea.KeyMsg:
		// 检查是否是Ctrl+C，如果是则允许打断处理
		if msg.String() == "ctrl+c" && m.isProcessing {
			m.isProcessing = false
			m.wasInterrupted = true
			// 清理打字机状态和临时内容
			m.streamingMsg = ""
			m.typingFull = ""
			m.typingPos = 0
			return m, nil
		}

		if m.isProcessing {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			if m.ctrlCPressed {
				// 第二次按Ctrl+C，完全退出
				return m, tea.Quit
			} else {
				// 第一次按Ctrl+C，取消当前输入
				m.input = ""
				m.cursor = 0
				m.ctrlCPressed = true
				// 不设置wasInterrupted，因为只是清空输入
				return m, nil
			}
		case "enter":
			if strings.TrimSpace(m.input) != "" {
				// 添加用户消息
				m.messages = append(m.messages, Message{
					Content: m.input,
					Sender:  "user",
					Time:    time.Now(),
					Type:    "text",
				})
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()

				// 开始处理
				m.isProcessing = true
				m.processingText = "思考中"
				inputContent := m.input
				m.input = ""
				m.cursor = 0             // 重置光标位置
				m.ctrlCPressed = false   // 重置Ctrl+C状态
				m.wasInterrupted = false // 重置中断状态
				m.streamingMsg = ""      // 重置流式输出
				m.typingFull = ""
				m.typingPos = 0

				return m, tea.Batch(m.processJavaLog(inputContent), m.updateProcessingText())
			}
		case "backspace":
			if len(m.input) > 0 && m.cursor > 0 {
				// 按字符删除，正确处理多字节UTF-8字符
				runes := []rune(m.input)
				if m.cursor <= len(runes) {
					// 删除光标前面的字符
					runes = append(runes[:m.cursor-1], runes[m.cursor:]...)
					m.input = string(runes)
					m.cursor-- // 光标向前移动
				}
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			runes := []rune(m.input)
			if m.cursor < len(runes) {
				m.cursor++
			}
		case "home":
			// 移动到行首
			m.cursor = 0
		case "end":
			// 移动到行尾
			runes := []rune(m.input)
			m.cursor = len(runes)
		case "delete":
			// 删除光标后面的字符
			runes := []rune(m.input)
			if m.cursor < len(runes) {
				runes = append(runes[:m.cursor], runes[m.cursor+1:]...)
				m.input = string(runes)
			}
		default:
			// 处理UTF-8字符输入
			if isValidInput(msg.String()) {
				// 在光标位置插入字符
				runes := []rune(m.input)
				if m.cursor <= len(runes) {
					runes = append(runes[:m.cursor], append([]rune(msg.String()), runes[m.cursor:]...)...)
					m.input = string(runes)
					m.cursor += len([]rune(msg.String())) // 向前移动光标
				}
				m.ctrlCPressed = false // 输入时重置Ctrl+C状态
			}
		}

	case AnalysisCompleteMsg:
		m.isProcessing = false
		m.wasInterrupted = false // 重置中断状态
		if msg.Error != nil {
			m.messages = append(m.messages, Message{
				Content: fmt.Sprintf("❌ 背景分析失败: %v\n\n现在您可以开始聊天了。", msg.Error),
				Sender:  "bot",
				Time:    time.Now(),
				Type:    "error",
			})
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case StreamMsg:
		if msg.Error != nil {
			m.isProcessing = false
			m.messages = append(m.messages, Message{
				Content: fmt.Sprintf("❌ 分析出错: %v", msg.Error),
				Sender:  "bot",
				Time:    time.Now(),
				Type:    "error",
			})
			m.streamingMsg = ""
		} else if msg.Done {
			// 流式输出完成
			m.isProcessing = false
			if m.streamingMsg != "" {
				m.messages = append(m.messages, Message{
					Content: m.streamingMsg,
					Sender:  "bot",
					Time:    time.Now(),
					Type:    "analysis",
				})
				m.streamingMsg = ""
			}
			if m.isFirst {
				return m, func() tea.Msg {
					return analysisDoneMsg{}
				}
			}
		} else {
			// 更新流式输出内容
			m.streamingMsg += msg.Content
			// 继续读取下一个流式数据块
			return m, m.continueStreaming()
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case tea.MouseMsg:
		// 允许使用鼠标滚轮
		if msg.Type == tea.MouseWheelUp {
			m.viewport.LineUp(1)
		} else if msg.Type == tea.MouseWheelDown {
			m.viewport.LineDown(1)
		}

	case processingTickMsg:
		if m.isProcessing {
			baseText := "思考中"
			numDots := (strings.Count(m.processingText, ".") + 1) % 4
			m.processingText = baseText + strings.Repeat(".", numDots)
			return m, m.updateProcessingText()
		}

	case typingTickMsg:
		// 打字机效果：每次追加少量rune
		if !m.isProcessing {
			return m, nil
		}
		runes := []rune(m.typingFull)
		if m.typingPos >= len(runes) {
			// 完成：将临时内容固化为消息
			if m.streamingMsg != "" {
				m.messages = append(m.messages, Message{
					Content: m.streamingMsg,
					Sender:  "bot",
					Time:    time.Now(),
					Type:    "analysis",
				})
				m.streamingMsg = ""
			}
			m.typingFull = ""
			m.typingPos = 0
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			if m.isFirst {
				return m, func() tea.Msg {
					return analysisDoneMsg{}
				}
			}
			m.isProcessing = false
			return m, nil
		}
		// 每tick输出3个rune（多语言友好）
		step := 3
		next := m.typingPos + step
		if next > len(runes) {
			next = len(runes)
		}
		m.streamingMsg += string(runes[m.typingPos:next])
		m.typingPos = next
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		// 安排下一次tick
		return m, tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg { return typingTickMsg(t) })
	case analysisDoneMsg:
		m.isProcessing = false
		m.wasInterrupted = false
		m.messages = append(m.messages, Message{
			Content: "✅ 背景分析完成！现在您可以开始聊天了。",
			Sender:  "bot",
			Time:    time.Now(),
			Type:    "text",
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	default:
		// 处理其他消息
		if m.isProcessing {
			return m, m.updateProcessingText()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ChatModel) View() string {
	if !m.ready {
		return "Initializing..."
	}
	var s strings.Builder

	// 标题样式
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.viewport.Width)
	s.WriteString(titleStyle.Render("🤖 Java启动分析器 - 交互式模式"))
	s.WriteString("\n" + strings.Repeat("─", m.viewport.Width) + "\n\n")

	// 显示消息历史
	s.WriteString(m.viewport.View())

	// 显示流式输出内容
	if m.streamingMsg != "" {
		streamingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
		timeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		timeStr := timeStyle.Render(time.Now().Format("15:04:05"))
		s.WriteString(streamingStyle.Render("🤖 分析器 ") + timeStr + "\n")
		s.WriteString(m.streamingMsg)
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// 显示处理状态
	if m.isProcessing && m.streamingMsg == "" {
		processingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true)
		s.WriteString(processingStyle.Render("⏳ " + m.processingText + "\n\n"))
	}

	// 分隔线
	s.WriteString(strings.Repeat("─", m.viewport.Width) + "\n")

	// 显示输入框
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	// 只有在真正打断处理时才显示红色提示
	if m.wasInterrupted {
		interruptedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // 红色
			Bold(true)
		s.WriteString(interruptedStyle.Render("💬 请输入您的问题 (Ctrl+C取消输入/退出, 方向键移动光标, Backspace/Delete删除): "))
	} else {
		s.WriteString(inputStyle.Render("💬 请输入您的问题 (Ctrl+C取消输入/退出, 方向键移动光标, Backspace/Delete删除): "))
	}

	// 显示输入内容，光标用下划线显示（不占用字符位置）
	runes := []rune(m.input)
	if len(runes) == 0 {
		// 空输入时显示光标
		cursorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Underline(true)
		s.WriteString(cursorStyle.Render(" "))
	} else if m.cursor >= len(runes) {
		// 光标在末尾
		s.WriteString(m.input)
		cursorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Underline(true)
		s.WriteString(cursorStyle.Render(" "))
	} else {
		// 光标在中间，下划线当前字符
		beforeCursor := string(runes[:m.cursor])
		currentChar := string(runes[m.cursor])
		afterCursor := string(runes[m.cursor+1:])

		s.WriteString(beforeCursor)

		// 下划线当前字符作为光标指示
		cursorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Underline(true)
		s.WriteString(cursorStyle.Render(currentChar))

		s.WriteString(afterCursor)
	}

	return s.String()
}

func (m *ChatModel) updateProcessingText() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return processingTickMsg(t)
	})
}

// startStreaming 启动真正的流式处理
func (m *ChatModel) startStreaming(streamReader *schema.StreamReader[*schema.Message]) tea.Cmd {
	return func() tea.Msg {
		// 读取流式数据
		message, err := streamReader.Recv()
		if err != nil {
			if err == io.EOF {
				// 流式输出完成
				return StreamMsg{Done: true}
			}
			// 发生错误
			return StreamMsg{Error: err, Done: true}
		}

		// 返回流式内容
		return StreamMsg{Content: message.Content, Done: false}
	}
}

// continueStreaming 继续流式处理
func (m *ChatModel) continueStreaming() tea.Cmd {
	return func() tea.Msg {
		if m.streamReader == nil {
			return StreamMsg{Error: fmt.Errorf("stream reader is nil"), Done: true}
		}

		// 读取下一个流式数据块
		message, err := m.streamReader.Recv()
		if err != nil {
			if err == io.EOF {
				// 流式输出完成
				return StreamMsg{Done: true}
			}
			// 发生错误
			return StreamMsg{Error: err, Done: true}
		}

		// 返回流式内容
		return StreamMsg{Content: message.Content, Done: false}
	}
}

func (m ChatModel) renderMessages() string {
	var s strings.Builder
	for _, msg := range m.messages {
		s.WriteString(m.renderMessage(msg))
		s.WriteString("\n")
	}
	return s.String()
}

func (m ChatModel) renderMessage(msg Message) string {
	var content strings.Builder

	// 时间戳
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	timeStr := timeStyle.Render(msg.Time.Format("15:04:05"))

	if msg.Sender == "user" {
		// 用户消息样式 - 简化显示
		userStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true)
		content.WriteString(userStyle.Render("👤 您 ") + timeStr + "\n")
		content.WriteString(msg.Content + "\n")
	} else {
		// 机器人消息样式 - 简化显示
		botStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
		content.WriteString(botStyle.Render("🤖 分析器 ") + timeStr + "\n")

		// 根据消息类型使用不同样式
		switch msg.Type {
		case "error":
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
			content.WriteString(errorStyle.Render(msg.Content))
		case "analysis":
			// 对于分析结果，只显示内容，不显示额外的格式
			content.WriteString(msg.Content)
		default:
			content.WriteString(msg.Content)
		}
	}

	return content.String()
}

// isValidInput 检查输入是否为有效的可打印字符（包括中文等多字节UTF-8字符）
func isValidInput(s string) bool {
	if s == "" {
		return false
	}

	// 检查是否包含可打印字符
	for _, r := range s {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return false
		}
	}

	// 过滤掉明显的控制字符和特殊键
	// 这些通常以特殊字符开头或包含控制字符
	if len(s) == 1 {
		// 单字符检查：排除控制字符
		if s[0] < 32 || s[0] == 127 {
			return false
		}
	}

	// 过滤掉常见的特殊键字符串（但保留方向键，因为它们在Update中单独处理）
	specialKeys := []string{
		"up", "down", "home", "end", "pageup", "pagedown",
		"ctrl+", "alt+", "shift+", "tab", "esc", "f1", "f2", "f3", "f4",
		"f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
	}

	for _, key := range specialKeys {
		if s == key || strings.HasPrefix(s, key) {
			return false
		}
	}

	return true
}

func (m ChatModel) processJavaLog(input string) tea.Cmd {
	return func() tea.Msg {
		// 构建消息
		ctx := context.Background()
		streamReader, err := m.analyzer.ChatStream(ctx, map[string]any{"input": input})
		if err != nil {
			return StreamMsg{Error: err, Done: true}
		}

		// 启动真正的流式处理
		return StartStreamMsg{StreamReader: streamReader, isFirst: false}
	}
}

func (m ChatModel) autoAnalyze() tea.Cmd {
	return func() tea.Msg {
		// 获取日志文件路径
		logPath, err := m.getLogFilePath()
		if err != nil {
			return AnalysisCompleteMsg{
				Error: fmt.Errorf("获取日志文件路径失败: %w", err),
			}
		}

		// 使用流式调用分析器，传递文件路径让大模型自己使用工具读取
		ctx := context.Background()
		streamReader, err := m.analyzer.ChatStream(ctx, map[string]any{"log_path": logPath})
		if err != nil {
			return AnalysisCompleteMsg{
				Error: err,
			}
		}

		// 启动流式处理
		return StartStreamMsg{StreamReader: streamReader, isFirst: true}
	}
}

func (m ChatModel) getLogFilePath() (string, error) {
	// 返回日志文件路径，让大模型自己使用工具读取
	if m.config.LogPath == "" {
		return "", fmt.Errorf("日志文件路径未配置")
	}
	return m.config.LogPath, nil
}

func (m ChatModel) readLogFile() (string, error) {
	// 读取日志文件
	content, err := os.ReadFile(m.config.LogPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

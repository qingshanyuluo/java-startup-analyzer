package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/java-startup-analyzer/internal/analyzer"
	"github.com/user/java-startup-analyzer/internal/ui"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "启动交互式聊天模式",
	Long: `启动交互式聊天模式，自动分析Java启动日志。

在聊天模式中，工具会：
- 自动读取配置文件中的启动命令和日志路径
- 自动开始分析Java启动日志
- 分析完成后允许您进行交互式聊天
- 获得智能的诊断和修复建议

使用 Ctrl+C 退出聊天模式。`,
	RunE: runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	// 检查配置文件是否指定
	if cfgFile == "" {
		return fmt.Errorf("请指定配置文件，使用 --config 参数")
	}

	// 创建分析器配置
	analyzerConfig := &analyzer.Config{
		Model:     viper.GetString("model"),
		ModelName: viper.GetString("model_name"),
		APIKey:    viper.GetString("api_key"),
		BaseURL:   viper.GetString("base_url"),
		Verbose:   viper.GetBool("verbose"),
		StartCmd:  viper.GetString("start_cmd"),
		LogPath:   viper.GetString("log_path"),
		GitRepo:   viper.GetString("git_repo"),
	}

	// 验证配置
	if err := analyzerConfig.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 创建聊天模型
	chatModel, err := ui.NewChatModel(analyzerConfig)
	if err != nil {
		return fmt.Errorf("创建聊天界面失败: %w", err)
	}

	// 启动Bubble Tea程序
	p := tea.NewProgram(chatModel, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		return fmt.Errorf("启动聊天界面失败: %w", err)
	}

	return nil
}

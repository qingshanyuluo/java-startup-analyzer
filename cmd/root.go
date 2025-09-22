package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "java-analyzer",
	Short: "Java程序启动失败分析工具",
	Long: `Java Startup Analyzer 是一个基于LLM的智能分析工具，
用于分析Java应用程序的启动日志，识别启动失败的原因并提供解决建议。

该工具使用Eino框架构建，支持多种LLM提供商，能够智能理解Java启动错误
并提供专业的诊断和修复建议。

所有配置（包括模型选择、API密钥等）都通过配置文件进行设置。
使用交互式聊天模式：java-analyzer chat --config config.yaml`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (必需)")
	rootCmd.PersistentFlags().String("model", "openai", "LLM模型提供商 (openai, anthropic, etc.)")
	rootCmd.PersistentFlags().String("api-key", "", "LLM API密钥")
	rootCmd.PersistentFlags().String("base-url", "", "LLM API基础URL")
	rootCmd.PersistentFlags().Bool("verbose", false, "详细输出模式")

	// 绑定标志到viper
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig 读取配置文件和环境变量
func initConfig() {
	if cfgFile != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 搜索配置文件
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".java-analyzer")
	}

	// 读取环境变量
	viper.AutomaticEnv()

	// 如果找到配置文件，读取它
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
		}
	}
}

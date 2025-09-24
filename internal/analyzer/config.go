package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config 分析器配置
type Config struct {
	Model     string // LLM模型提供商
	ModelName string // 具体模型名称 (如 gpt-4.1, gpt-3.5-turbo)
	APIKey    string // API密钥
	BaseURL   string // API基础URL
	Verbose   bool   // 详细输出模式
	StartCmd  string // 启动命令 (必需)
	LogPath   string // 日志文件路径 (必需)
	LogDir    string // 分析器日志目录 (可选，默认为 ./logs)
	GitRepo   string // Git仓库路径 (可选)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Model:     "openai",
		ModelName: "gpt-3.5-turbo",
		Verbose:   false,
		LogDir:    "./logs", // 默认日志目录
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("api密钥不能为空")
	}
	if c.StartCmd == "" {
		return fmt.Errorf("启动命令不能为空")
	}
	if c.LogPath == "" {
		return fmt.Errorf("日志路径不能为空")
	}

	// 将相对路径转换为绝对路径
	if !filepath.IsAbs(c.LogPath) {
		absLogPath, err := filepath.Abs(c.LogPath)
		if err != nil {
			return fmt.Errorf("无法解析日志文件路径: %w", err)
		}
		c.LogPath = absLogPath
	}

	// 检查日志文件是否存在
	if _, err := os.Stat(c.LogPath); os.IsNotExist(err) {
		return fmt.Errorf("日志文件不存在: %s", c.LogPath)
	}

	// 如果指定了Git仓库，检查是否存在
	if c.GitRepo != "" {
		// 将Git仓库相对路径转换为绝对路径
		if !filepath.IsAbs(c.GitRepo) {
			absGitPath, err := filepath.Abs(c.GitRepo)
			if err != nil {
				return fmt.Errorf("无法解析git仓库路径: %w", err)
			}
			c.GitRepo = absGitPath
		}

		gitPath := filepath.Join(c.GitRepo, ".git")
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			return fmt.Errorf("git仓库不存在: %s", c.GitRepo)
		}
	}

	return nil
}

package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// Client LLM客户端
type Client struct {
	model     model.ChatModel
	modelType string
	modelName string
	apiKey    string
	baseURL   string
}

// NewClient 创建新的LLM客户端
func NewClient(modelType, modelName, apiKey, baseURL string) (*Client, error) {
	client := &Client{
		modelType: modelType,
		modelName: modelName,
		apiKey:    apiKey,
		baseURL:   baseURL,
	}

	// 根据模型类型创建相应的模型实例
	var err error
	switch modelType {
	case "openai":
		client.model, err = createOpenAIModel(modelName, apiKey, baseURL)
	// case "anthropic":
	// 	client.model, err = createAnthropicModel(modelName, apiKey, baseURL)
	default:
		return nil, fmt.Errorf("不支持的模型类型: %s", modelType)
	}

	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	return client, nil
}

// GetChatModel 获取聊天模型
func (c *Client) GetChatModel() model.ChatModel {
	return c.model
}

// createOpenAIModel 创建OpenAI模型
func createOpenAIModel(modelName, apiKey, baseURL string) (model.ChatModel, error) {
	// 使用 Eino 官方的 OpenAI 实现
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 OpenAI 模型失败: %w", err)
	}
	return chatModel, nil
}

package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/sashabaranov/go-openai"
)

// OpenAIModel is a chat model implementation using the OpenAI API.
type OpenAIModel struct {
	client    *openai.Client
	modelName string
}

// NewOpenAIModel creates a new OpenAIModel.
func NewOpenAIModel(modelName, apiKey, baseURL string) (*OpenAIModel, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key cannot be empty")
	}

	// 如果没有指定模型名称，使用默认值
	if modelName == "" {
		modelName = openai.GPT3Dot5Turbo
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)
	return &OpenAIModel{
		client:    client,
		modelName: modelName,
	}, nil
}

// Generate generates a chat completion using the OpenAI API.
func (m *OpenAIModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	messages := make([]openai.ChatCompletionMessage, len(input))
	for i, msg := range input {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    m.modelName,
		Messages: messages,
	}

	resp, err := m.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := resp.Choices[0].Message
	return &schema.Message{
		Role:    schema.RoleType(choice.Role),
		Content: choice.Content,
	}, nil
}

// Stream is not yet implemented.
func (m *OpenAIModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("streaming is not yet implemented for the OpenAI model")
}

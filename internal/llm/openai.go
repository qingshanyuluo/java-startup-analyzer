package llm

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/sashabaranov/go-openai"
)

// OpenAIModel is a chat model implementation using the OpenAI API.
type OpenAIModel struct {
	client    *openai.Client
	modelName string
	tools     []*schema.ToolInfo
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
		tools:     []*schema.ToolInfo{},
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

	// 添加工具支持
	if len(m.tools) > 0 {
		req.Tools = make([]openai.Tool, len(m.tools))
		for i, tool := range m.tools {
			req.Tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Desc,
					Parameters:  tool.ParamsOneOf,
				},
			}
		}
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

// Stream implements streaming chat completion using the OpenAI API.
func (m *OpenAIModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
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
		Stream:   true,
	}

	stream, err := m.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}

	// 创建StreamReader和StreamWriter
	reader, writer := schema.Pipe[*schema.Message](10)

	// 启动goroutine处理流式响应
	go func() {
		defer writer.Close()
		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				// 发送错误到流
				writer.Send(nil, fmt.Errorf("failed to receive stream response: %w", err))
				return
			}

			if len(response.Choices) == 0 {
				// 流式响应的最后一个chunk可能没有choices，这是正常的
				continue
			}

			choice := response.Choices[0]

			// 检查是否有内容需要发送
			if choice.Delta.Content == "" && choice.Delta.Role == "" {
				// 空的delta，跳过
				continue
			}

			message := &schema.Message{
				Role:    schema.RoleType(choice.Delta.Role),
				Content: choice.Delta.Content,
			}

			// 发送消息到流
			closed := writer.Send(message, nil)
			if closed {
				break
			}
		}
	}()

	return reader, nil
}

// BindTools binds tools to the model.
func (m *OpenAIModel) BindTools(tools []*schema.ToolInfo) error {
	m.tools = tools
	return nil
}

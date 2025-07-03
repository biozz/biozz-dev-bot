package gpts

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
}

func NewOpenAIProvider(apiKey string) Provider {
	client := openai.NewClient(apiKey)
	return &Client{client: client}
}

func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	if req.System != "" {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    "system",
			Content: req.System,
		})
	}
	for _, m := range req.Messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    req.Model,
			Messages: openaiMessages,
		},
	)
	if err != nil {
		return ChatCompletionResponse{}, err
	}
	result := ChatCompletionResponse{
		Message: ChatCompletionMessage{
			Role:    resp.Choices[0].Message.Role,
			Content: resp.Choices[0].Message.Content,
		},
	}
	return result, nil
}

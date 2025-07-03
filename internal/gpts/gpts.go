package gpts

import (
	"context"
)

const (
	OpenAI     string = "openai"
	OpenRouter string = "openrouter"
)

const (
	GPT4o string = "gpt-4o"
	GPT4  string = "gpt-4"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleSystem    Role = "system"
	RoleAssistant Role = "assistant"
)

var Providers = map[string][]string{
	OpenAI: {GPT4o},
}

type Provider interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error)
}

func NewProvider(providerType string, apiKey string) Provider {
	switch providerType {
	case OpenAI:
		return NewOpenAIProvider(apiKey)
	default:
		return nil
	}
}

type ChatCompletionMessage struct {
	Role    string
	Content string
}

type ChatCompletionRequest struct {
	Model    string
	System   string
	Messages []ChatCompletionMessage
}

type ChatCompletionResponse struct {
	Message ChatCompletionMessage
}

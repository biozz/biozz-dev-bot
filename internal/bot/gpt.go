package bot

import (
	"context"
	"fmt"

	"github.com/biozz/biozz-dev-bot/internal/gpts"
	"github.com/biozz/biozz-dev-bot/internal/librechat"
	tele "gopkg.in/telebot.v4"
)

var (
	gptModelsMenu                        = &tele.ReplyMarkup{}
	libreChatProviders map[string]string = map[string]string{
		"openAI": gpts.OpenAI,
	}
)

func (b *Bot) newGPTChat(c tele.Context) error {

	convo, err := b.librechatClient.MongoCreateConversation(librechat.EndpointOpenAI)
	if err != nil {
		return err
	}

	err = b.setState(map[string]any{"convo": convo, "chat_state": "gpt"})
	if err != nil {
		return err
	}
	msg := "Started new conversation"
	return c.Send(msg, &tele.SendOptions{ReplyMarkup: gptModelsMenu})
}

func (b *Bot) handleGPTMessage(c tele.Context) error {
	var (
		txt = c.Text()
	)

	convoID, err := b.getState("convo")
	if err != nil {
		return c.Send("Unable to get conversation from DB")
	}

	convo, err := b.librechatClient.MongoGetConversation(convoID)
	if err != nil {
		return c.Send("Unable to get conversation from DB")
	}

	providerName, ok := libreChatProviders[convo.Endpoint]
	if !ok {
		return c.Send("Unable to match provider")
	}

	messages, err := b.librechatClient.MongoGetConversationMessages(convoID)
	if err != nil {
		return c.Send("Unable to get conversation messages from DB")
	}

	parentID := librechat.DefaultParentMessageID
	if len(messages) > 0 {
		parentID = messages[len(messages)-1].ID
	}

	lastUserMessageID, err := b.librechatClient.MongoCreateMessage(convoID, txt, parentID, true)
	if err != nil {
		return c.Send("Unable to create message in DB")
	}

	var completionMessages []gpts.ChatCompletionMessage

	// Add system prompt for concise responses
	completionMessages = append(completionMessages, gpts.ChatCompletionMessage{
		Role:    "system",
		Content: "You are a helpful assistant. Keep your responses concise and to the point.",
	})

	for i := range messages {
		role := gpts.RoleAssistant
		if messages[i].IsCreatedByUser {
			role = gpts.RoleUser
		}
		completionMessages = append(completionMessages, gpts.ChatCompletionMessage{
			Role:    string(role),
			Content: messages[i].Text,
		})
	}

	// Add current user message
	completionMessages = append(completionMessages, gpts.ChatCompletionMessage{
		Role:    string(gpts.RoleUser),
		Content: txt,
	})

	c.Notify(tele.Typing)

	// TODO: change to dynamic API key
	provider := gpts.NewProvider(providerName, b.openAIAPIKey)

	resp, err := provider.CreateChatCompletion(
		context.Background(),
		gpts.ChatCompletionRequest{
			Model:    convo.Model,
			Messages: completionMessages,
		},
	)
	if err != nil {
		return c.Send(fmt.Sprintf("chat completion error: %v", err))
	}

	result := resp.Message.Content

	b.librechatClient.MongoCreateMessage(convoID, result, lastUserMessageID, false)

	// If this is the first response (only 2 messages: user question + GPT response)
	// Generate a summary title using o3-mini
	if len(messages) == 0 {
		go b.summarizeConversation(convoID, txt, result)
	}

	c.Send(result)

	return nil

}

func (b *Bot) summarizeConversation(convoID string, userMessage string, gptResponse string) {
	provider := gpts.NewProvider(gpts.OpenAI, b.openAIAPIKey)

	summaryPrompt := fmt.Sprintf(`Generate a concise title (max 4-5 words) for this conversation based on the user's question and assistant's response:

User: %s
Assistant: %s

Title:`, userMessage, gptResponse)

	resp, err := provider.CreateChatCompletion(
		context.Background(),
		gpts.ChatCompletionRequest{
			Model: b.summaryModel,
			Messages: []gpts.ChatCompletionMessage{
				{
					Role:    string(gpts.RoleUser),
					Content: summaryPrompt,
				},
			},
		},
	)
	if err != nil {
		return // Fail silently, keep original title
	}

	title := resp.Message.Content
	b.librechatClient.MongoUpdateConversationTitle(convoID, title)
}

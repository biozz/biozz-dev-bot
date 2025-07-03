package bot

import (
	"context"
	"fmt"

	"github.com/biozz/biozz-dev-bot/internal/gpts"
	"github.com/biozz/biozz-dev-bot/internal/librechat"
	"github.com/pocketbase/pocketbase/core"
	"gopkg.in/telebot.v4"
	tele "gopkg.in/telebot.v4"
)

var (
	gptModelsMenu                        = &tele.ReplyMarkup{}
	openAIGPT41                          = gptModelsMenu.Data("OpenAI – GPT-4o", "openAI/gpt-4o")
	libreChatProviders map[string]string = map[string]string{
		"openAI": gpts.OpenAI,
	}
)

func (b *Bot) newGPTChat(c tele.Context) error {

	convo, err := b.librechatClient.MongoCreateConversation(
		"67d54d915f5f1dfc7994e482",
		"openAI",
		"gpt-4o",
	)
	if err != nil {
		return err
	}

	user := c.Get("user").(*core.Record)
	err = b.setUserData(user, map[string]any{"convo": convo, "chat_state": "gpt"})
	if err != nil {
		return err
	}
	msg := "Starting new conversation with OpenAI GPT-4o, change it if necessary"
	msg = EscapeTelegramMarkdown(msg)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdownV2}, &gptModelsMenu)
}

func (b *Bot) handleGPTMessage(c tele.Context) error {
	var (
		txt  = c.Text()
		user = c.Get("user").(*core.Record)
	)

	convoID := user.Get("convo").(string)

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

	completionMessages := make([]gpts.ChatCompletionMessage, len(messages))
	for i := range messages {
		role := gpts.RoleAssistant
		if messages[i].IsCreatedByUser {
			role = gpts.RoleUser
		}
		completionMessages[i] = gpts.ChatCompletionMessage{
			Role:    string(role),
			Content: messages[i].Text,
		}
	}

	c.Notify(telebot.Typing)

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

	c.Send(result)

	return nil

}

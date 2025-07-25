package bot

import (
	"strings"
	"time"

	ha "github.com/biozz/biozz-dev-bot/internal/homeassistant"
	"github.com/biozz/biozz-dev-bot/internal/librechat"
	"github.com/pocketbase/pocketbase"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

type Bot struct {
	app              *pocketbase.PocketBase
	bot              *tele.Bot
	librechatClient  *librechat.LibreChat
	haClient         *ha.HomeAssistant
	superuserID      int64
	supergroupID     int64
	gptThreadID      int64
	openAIAPIKey     string
	openRouterAPIKey string
	summaryModel     string
}

type NewBotParams struct {
	App                 *pocketbase.PocketBase
	LibreChatClient     *librechat.LibreChat
	HomeAssistantClient *ha.HomeAssistant
	BotToken            string
	SuperGroupID        int64
	SuperUserID         int64
	GPTThreadID         int64
	// API Keys
	OpenAIAPIKey     string
	OpenRouterAPIKey string
	SummaryModel     string
}

func New(params NewBotParams) (*Bot, error) {
	pref := tele.Settings{
		Token: params.BotToken,
		Poller: &tele.LongPoller{
			Timeout: 10 * time.Second,
			AllowedUpdates: []string{
				"message",
				"callback_query",
			},
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		bot:             b,
		librechatClient: params.LibreChatClient,
		haClient:        params.HomeAssistantClient,
		app:             params.App,
		superuserID:     params.SuperUserID,
		supergroupID:    params.SuperGroupID,
		gptThreadID:     params.GPTThreadID,
		// API Keys
		openAIAPIKey:     params.OpenAIAPIKey,
		openRouterAPIKey: params.OpenRouterAPIKey,
		summaryModel:     params.SummaryModel,
	}
	return bot, nil
}

func (b *Bot) Start() {

	// Register global middleware
	b.bot.Use(b.LogMessage)
	b.bot.Use(middleware.Whitelist(b.superuserID))

	// Main commands
	b.bot.Handle("/gpt", b.newGPTChat)
	b.bot.Handle("/ha", b.handleHomeAssistant)

	b.bot.Handle(tele.OnCallback, b.handleCallback)
	b.bot.Handle(tele.OnText, b.handleText)

	b.bot.Start()
}

func (b *Bot) handleCallback(c tele.Context) error {
	data := c.Callback().Data
	data = strings.TrimLeft(data, "\f")
	b.app.Logger().Debug("Received callback", "data", data)

	// Handle Home Assistant device callbacks
	if strings.HasPrefix(data, "ha:") {
		return b.handleHomeAssistantCallback(c)
	}

	return nil
}

func (b *Bot) handleText(c tele.Context) error {
	state, err := b.getState("chat_state")
	if err != nil {
		b.app.Logger().Error("Error getting state", "error", err)
		return err
	}
	b.app.Logger().Debug("Received text", "text", c.Text(), "state", state)

	if c.Message().ThreadID == int(b.gptThreadID) {
		b.app.Logger().Debug("Received text", "text", c.Text(), "state", state)
		return b.handleGPTMessage(c)
	}

	return nil
}

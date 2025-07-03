package bot

import (
	"strings"
	"time"

	"github.com/biozz/biozz-dev-bot/internal/librechat"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

type Bot struct {
	app              *pocketbase.PocketBase
	bot              *tele.Bot
	librechatClient  *librechat.LibreChat
	superuserID      int64
	supergroupID     int64
	gptThreadID      int64
	openAIAPIKey     string
	openRouterAPIKey string
}

type NewBotParams struct {
	App             *pocketbase.PocketBase
	LibreChatClient *librechat.LibreChat
	BotToken        string
	SuperGroupID    int64
	SuperUserID     int64
	GPTThreadID     int64
	// API Keys
	OpenAIAPIKey     string
	OpenRouterAPIKey string
}

func New(params NewBotParams) (*Bot, error) {
	pref := tele.Settings{
		Token:  params.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		bot:             b,
		librechatClient: params.LibreChatClient,
		app:             params.App,
		superuserID:     params.SuperUserID,
		supergroupID:    params.SuperGroupID,
		gptThreadID:     params.GPTThreadID,
		// API Keys
		openAIAPIKey:     params.OpenAIAPIKey,
		openRouterAPIKey: params.OpenRouterAPIKey,
	}
	return bot, nil
}

func (b *Bot) Start() {

	// Register global middleware
	b.bot.Use(b.SetUser)
	b.bot.Use(b.LogMessage)
	b.bot.Use(middleware.Whitelist(b.superuserID))

	// Main commands
	b.bot.Handle("/start", b.handleStart)

	b.bot.Handle(tele.OnCallback, b.handleCallback)
	b.bot.Handle(tele.OnText, b.handleText)

	b.bot.Start()
}

func (b *Bot) handleCallback(c tele.Context) error {
	data := c.Callback().Data
	data = strings.TrimLeft(data, "\f")
	b.app.Logger().Debug("Received callback", "data", data)
	return nil
}

func (b *Bot) handleText(c tele.Context) error {
	user := c.Get("user").(*core.Record)
	state := user.GetString("chat_state")
	b.app.Logger().Debug("Received text", "text", c.Text(), "state", state)

	switch state {
	case "gpt":
		b.app.Logger().Debug("Received text", "text", c.Text(), "state", state)
		return b.handleGPTMessage(c)
	}

	return nil
}

func (b *Bot) handleStart(c tele.Context) error {
	return nil
}

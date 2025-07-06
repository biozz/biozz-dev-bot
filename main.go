package main

import (
	"log"
	"os"
	"strings"

	_ "github.com/biozz/biozz-dev-bot/migrations"

	"github.com/biozz/biozz-dev-bot/internal/bot"
	ha "github.com/biozz/biozz-dev-bot/internal/homeassistant"
	"github.com/biozz/biozz-dev-bot/internal/librechat"
	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

type config struct {
	TelegramBotToken   string `env:"TELEGRAM_BOT_TOKEN"`
	SuperGroupID       int64  `env:"SUPERGROUP_ID"`
	SuperUserID        int64  `env:"SUPERUSER_ID"`
	GPTThreadID        int64  `env:"GPT_THREAD_ID"`
	LibreChatMongoURI  string `env:"LIBRECHAT_MONGO_URI"`
	LibreChatUserID    string `env:"LIBRECHAT_USER_ID"`
	LibreChatTag       string `env:"LIBRECHAT_TAG"`
	OpenAIAPIKey       string `env:"OPENAI_API_KEY"`
	OpenRouterAPIKey   string `env:"OPENROUTER_API_KEY"`
	ConvoModel         string `env:"CONVO_MODEL"`
	SummaryModel       string `env:"SUMMARY_MODEL"`
	HomeAssistantURL   string `env:"HOME_ASSISTANT_URL"`
	HomeAssistantToken string `env:"HOME_ASSISTANT_TOKEN"`
}

func main() {
	app := pocketbase.New()

	cfg, err := env.ParseAs[config]()
	if err != nil {
		app.Logger().Error("Failed to parse environment variables", "error", err)
		return
	}

	librechatClient, err := librechat.New(librechat.NewParams{
		MongoURI:    cfg.LibreChatMongoURI,
		MongoUserID: cfg.LibreChatUserID,
		MongoTag:    cfg.LibreChatTag,
		ConvoModel:  cfg.ConvoModel,
	})
	if err != nil {
		app.Logger().Error("Failed to create LibreChat client", "error", err)
		return
	}
	defer librechatClient.Cleanup()

	haClient := ha.New(cfg.HomeAssistantURL, cfg.HomeAssistantToken)

	bot, err := bot.New(bot.NewBotParams{
		App:                 app,
		LibreChatClient:     librechatClient,
		HomeAssistantClient: haClient,
		BotToken:            cfg.TelegramBotToken,
		SuperGroupID:        cfg.SuperGroupID,
		SuperUserID:         cfg.SuperUserID,
		GPTThreadID:         cfg.GPTThreadID,
		OpenAIAPIKey:        cfg.OpenAIAPIKey,
		OpenRouterAPIKey:    cfg.OpenRouterAPIKey,
		SummaryModel:        cfg.SummaryModel,
	})
	if err != nil {
		app.Logger().Error("Failed to create bot", "error", err)
		return
	}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		go bot.Start()

		return se.Next()
	})

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
		Dir:         "./migrations",
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

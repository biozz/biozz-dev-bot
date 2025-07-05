package bot

import (
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) LogMessage(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.app.Logger().Info(
			"New message",
			"text", c.Message().Text,
			"chat_id", c.Chat().ID,
			"user_id", c.Sender().ID,
			"thread_id", c.Message().ThreadID,
		)
		return next(c)
	}
}

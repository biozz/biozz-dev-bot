package bot

import (
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) SetUser(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, _ := b.app.FindFirstRecordByData("users", "chat_id", c.Sender().ID)
		c.Set("user", user)
		return next(c)
	}
}

func (b *Bot) LogMessage(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		// We don't want to log all the messages, just the private chat with the bot
		if c.Chat().Type != tele.ChatPrivate {
			return nil
		}
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

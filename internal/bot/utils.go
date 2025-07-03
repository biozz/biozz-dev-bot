package bot

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func EscapeTelegramMarkdown(text string) string {
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, ".", "\\.")
	text = strings.ReplaceAll(text, "-", "\\-")
	text = strings.ReplaceAll(text, "*", "\\*")
	text = strings.ReplaceAll(text, "[", "\\[")
	return text
}

func (b *Bot) setUserData(user *core.Record, data map[string]any) error {
	for key, value := range data {
		user.Set(key, value)
	}
	return b.app.Save(user)
}

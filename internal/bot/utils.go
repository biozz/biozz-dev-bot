package bot

import (
	"strings"

	"github.com/pocketbase/dbx"
)

func EscapeTelegramMarkdown(text string) string {
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, ".", "\\.")
	text = strings.ReplaceAll(text, "-", "\\-")
	text = strings.ReplaceAll(text, "*", "\\*")
	text = strings.ReplaceAll(text, "[", "\\[")
	return text
}

func (b *Bot) setState(data map[string]any) error {
	for key, value := range data {
		state, err := b.app.FindFirstRecordByFilter("state", "key = {:key}", dbx.Params{"key": key})
		if err != nil {
			b.app.Logger().Error("Error finding state", "error", err)
			return err
		}
		state.Set("value", value)
		if err := b.app.Save(state); err != nil {
			b.app.Logger().Error("Error saving state", "error", err)
			return err
		}
	}
	return nil
}

func (b *Bot) getState(key string) (string, error) {
	state, err := b.app.FindFirstRecordByFilter("state", "key = {:key}", dbx.Params{"key": key})
	if err != nil {
		b.app.Logger().Error("Error finding state", "error", err)
		return "", err
	}
	return state.Get("value").(string), nil
}

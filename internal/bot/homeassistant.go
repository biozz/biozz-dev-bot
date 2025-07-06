package bot

import (
	"fmt"
	"strings"

	ha "github.com/biozz/biozz-dev-bot/internal/homeassistant"
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) handleHomeAssistant(c tele.Context) error {
	// Get devices from pocketbase
	devices, err := b.getDevices()
	if err != nil {
		b.app.Logger().Error("Error getting devices", "error", err)
		return c.Reply("❌ Error getting devices")
	}

	if len(devices) == 0 {
		return c.Reply("📱 No devices found in database")
	}

	// Create keyboard with devices and refresh button
	keyboard := b.createDeviceKeyboard(devices)

	return c.Send("🏠 Home Assistant Devices:", keyboard)
}

func (b *Bot) createDeviceKeyboard(devices []ha.Device) *tele.ReplyMarkup {
	keyboard := &tele.ReplyMarkup{}
	var rows []tele.Row

	// Add device buttons
	for _, device := range devices {
		btn := keyboard.Data(
			fmt.Sprintf("%s %s", getDeviceIcon(device.Type), device.Name),
			fmt.Sprintf("ha:%s", device.EntityID),
		)
		rows = append(rows, keyboard.Row(btn))
	}

	// Add refresh button as the last button
	refreshBtn := keyboard.Data("🔄 Refresh", "ha:refresh")
	rows = append(rows, keyboard.Row(refreshBtn))

	keyboard.Inline(rows...)
	return keyboard
}

func (b *Bot) handleHomeAssistantCallback(c tele.Context) error {
	data := c.Callback().Data
	entityID := strings.TrimPrefix(data, "\fha:")

	b.app.Logger().Debug("Home Assistant callback", "entity_id", entityID)

	// Handle refresh action
	if entityID == "refresh" {
		// Get updated devices from database
		devices, err := b.getDevices()
		if err != nil {
			b.app.Logger().Error("Error getting devices for refresh", "error", err)
			return c.Respond(&tele.CallbackResponse{Text: "❌ Error refreshing devices"})
		}

		if len(devices) == 0 {
			return c.Respond(&tele.CallbackResponse{Text: "📱 No devices found"})
		}

		// Create new keyboard with updated devices
		keyboard := b.createDeviceKeyboard(devices)

		// Edit the message with the new keyboard
		err = c.Edit("🏠 Home Assistant Devices:", keyboard)
		if err != nil {
			b.app.Logger().Error("Error editing message", "error", err)
			return c.Respond(&tele.CallbackResponse{Text: "❌ Error refreshing"})
		}

		return c.Respond(&tele.CallbackResponse{Text: "✅ Devices refreshed"})
	}

	// Get device from database to determine action
	device, err := b.getDevice(entityID)
	if err != nil {
		b.app.Logger().Error("Error getting device", "error", err)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Device not found"})
	}

	// Perform action based on device type
	var action string

	switch device.Type {
	case "light", "switch":
		// For lights and switches, toggle them
		err = b.haClient.Toggle(entityID)
		action = "toggled"
	case "button":
		// For buttons, press them
		err = b.haClient.PressButton(entityID)
		action = "pressed"
	default:
		// For other devices, try to toggle
		err = b.haClient.Toggle(entityID)
		action = "toggled"
	}

	if err != nil {
		b.app.Logger().Error("Error controlling device", "error", err, "entity_id", entityID)
		return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("❌ Failed to control %s", device.Name)})
	}

	return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("✅ %s %s", device.Name, action)})
}

func (b *Bot) getDevices() ([]ha.Device, error) {
	records, err := b.app.FindRecordsByFilter("devices", "", "-created", 0, 0)
	if err != nil {
		return nil, err
	}

	var devices []ha.Device
	for _, record := range records {
		deviceType := strings.Split(record.GetString("entity_id"), ".")[0]
		device := ha.Device{
			EntityID: record.GetString("entity_id"),
			Name:     record.GetString("name"),
			Type:     deviceType,
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (b *Bot) getDevice(entityID string) (*ha.Device, error) {
	record, err := b.app.FindFirstRecordByFilter("devices", "entity_id = {:entityID}", map[string]any{
		"entityID": entityID,
	})
	if err != nil {
		return nil, err
	}

	deviceType := strings.Split(record.GetString("entity_id"), ".")[0]

	device := &ha.Device{
		EntityID: record.GetString("entity_id"),
		Name:     record.GetString("name"),
		Type:     deviceType,
	}

	return device, nil
}

func getDeviceIcon(deviceType string) string {
	switch deviceType {
	case "light":
		return "💡"
	case "switch":
		return "🔌"
	case "button":
		return "🔘"
	default:
		return "📱"
	}
}

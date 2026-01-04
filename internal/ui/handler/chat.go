package handler

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgChat(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.ChatPayload
	if err := payloadconv.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	sender := payload.SenderName
	if sender == "" {
		sender = "未知"
	}

	timeStr := time.Unix(payload.Time, 0).Format("15:04")
	chatLine := fmt.Sprintf("[%s] %s: %s", timeStr, sender, payload.Content)
	if payload.IsSystem {
		chatLine = fmt.Sprintf("[%s] 系统: %s", timeStr, payload.Content)
	}

	// Route message to appropriate chat based on scope
	switch payload.Scope {
	case "lobby":
		m.Lobby().AddChatMessage(chatLine)
	case "room":
		m.Game().AddChatMessage(chatLine)
	default:
		// Fallback: add to current context
		if m.Game().State().RoomCode != "" {
			m.Game().AddChatMessage(chatLine)
		} else {
			m.Lobby().AddChatMessage(chatLine)
		}
	}

	return nil
}

// setMaintenanceNotification 设置维护模式通知
func setMaintenanceNotification(m model.Model, maintenance bool) {
	m.SetMaintenanceMode(maintenance)
	if maintenance {
		m.SetNotification(model.NotifyMaintenance, "⚠️ 服务器维护中，暂停接受新连接", false)
	} else {
		m.ClearNotification(model.NotifyMaintenance)
	}
}

func handleMsgMaintenancePush(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.MaintenancePayload
	if err := payloadconv.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	setMaintenanceNotification(m, payload.Maintenance)
	return nil
}

func handleMsgMaintenancePull(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.MaintenanceStatusPayload
	if err := payloadconv.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	setMaintenanceNotification(m, payload.Maintenance)
	return nil
}

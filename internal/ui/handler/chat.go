package handler

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgChat(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.ChatPayload
	if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
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

	m.Lobby().AddChatMessage(chatLine)
	if m.Game().State().RoomCode != "" {
		m.Game().AddChatMessage(chatLine)
	}

	return nil
}

func handleMsgMaintenance(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.MaintenancePayload
	if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	m.SetMaintenanceMode(payload.Maintenance)
	if payload.Maintenance {
		m.SetNotification(model.NotifyMaintenance, "⚠️ 服务器维护中，暂停接受新连接", false)
	} else {
		m.ClearNotification(model.NotifyMaintenance)
	}

	return nil
}

func handleMsgMaintenanceStatus(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.MaintenanceStatusPayload
	if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	m.SetMaintenanceMode(payload.Maintenance)
	return nil
}

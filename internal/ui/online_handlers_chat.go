package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// handleMsgChat handles incoming chat messages
func (m *OnlineModel) handleMsgChat(msg *protocol.Message) tea.Cmd {
	var payload protocol.ChatPayload
	if err := protocol.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	sender := payload.SenderName
	if sender == "" {
		sender = "未知"
	}

	// Format: [12:30] Player: Message content
	timeStr := time.Unix(payload.Time, 0).Format("15:04")
	chatLine := fmt.Sprintf("[%s] %s: %s", timeStr, sender, payload.Content)
	if payload.IsSystem {
		chatLine = fmt.Sprintf("[%s] 系统: %s", timeStr, payload.Content)
	}

	// Update Lobby chat history
	m.lobby.chatHistory = append(m.lobby.chatHistory, chatLine)
	if len(m.lobby.chatHistory) > 50 {
		m.lobby.chatHistory = m.lobby.chatHistory[len(m.lobby.chatHistory)-50:]
	}

	// Update Game chat history if in game/room
	if m.game.roomCode != "" {
		m.game.chatHistory = append(m.game.chatHistory, chatLine)
		if len(m.game.chatHistory) > 50 {
			m.game.chatHistory = m.game.chatHistory[len(m.game.chatHistory)-50:]
		}
	}

	return nil
}

// handleMsgMaintenance handles server maintenance notifications
func (m *OnlineModel) handleMsgMaintenance(msg *protocol.Message) tea.Cmd {
	var payload protocol.MaintenancePayload
	if err := protocol.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	// Set maintenance mode flag
	m.maintenanceMode = payload.Maintenance

	// Display maintenance message as an error
	if payload.Maintenance {
		m.error = "⚠️  服务器正在维护中"
	} else {
		m.error = ""
	}

	return nil
}

// handleMsgMaintenanceStatus handles maintenance status response
func (m *OnlineModel) handleMsgMaintenanceStatus(msg *protocol.Message) tea.Cmd {
	var payload protocol.MaintenanceStatusPayload
	if err := protocol.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	// Update maintenance mode flag
	m.maintenanceMode = payload.Maintenance

	return nil
}

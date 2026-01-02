package model

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
)

func (m *OnlineModel) handleWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.lobby.SetSize(msg.Width, msg.Height)
	m.game.SetSize(msg.Width, msg.Height)
}

func (m *OnlineModel) handleConnected() tea.Cmd {
	m.EnterLobby()
	m.playerID = m.client.PlayerID
	m.playerName = m.client.PlayerName
	m.client.StartHeartbeat()
	return m.listenForMessages()
}

func (m *OnlineModel) handleConnectionError(msg ConnectionErrorMsg) {
	m.error = fmt.Sprintf("无法连接到服务器: %v\n\n按 ESC 退出", msg.Err)
	m.phase = PhaseConnecting
}

func (m *OnlineModel) handleReconnecting(msg ReconnectingMsg) tea.Cmd {
	m.reconnecting = true
	m.reconnectAttempt = msg.Attempt
	m.reconnectMaxTries = msg.MaxTries
	m.SetNotification(NotifyReconnecting, fmt.Sprintf("🔄 正在重连 (%d/%d)...", msg.Attempt, msg.MaxTries), false)
	return m.listenForReconnect()
}

func (m *OnlineModel) handleReconnectSuccess() []tea.Cmd {
	m.reconnecting = false
	m.ClearNotification(NotifyReconnecting)
	m.ClearNotification(NotifyError)
	m.ClearNotification(NotifyRateLimit)
	m.SetNotification(NotifyReconnectSuccess, "✅ 重连成功！", true)

	var cmds []tea.Cmd
	cmds = append(cmds, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ClearReconnectMsg{}
	}))
	cmds = append(cmds, m.listenForReconnect())

	if m.client.IsConnected() {
		cmds = append(cmds, m.listenForMessages())
	}
	return cmds
}

func (m *OnlineModel) handleClearReconnect() {
	m.ClearNotification(NotifyReconnectSuccess)
	if m.phase == PhaseLobby {
		_ = m.client.SendMessage(codec.MustNewMessage(protocol.MsgGetOnlineCount, nil))
		_ = m.client.SendMessage(codec.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))
	}
}

func (m *OnlineModel) handleClearInputError() {
	// Restore input placeholder after displaying error
	switch m.phase {
	case PhaseBidding:
		if m.game.BidTurn() == m.playerID {
			m.input.Placeholder = "叫地主? (Y/N)"
		}
	case PhasePlaying:
		if m.game.State().CurrentTurn == m.playerID {
			switch {
			case m.game.MustPlay():
				m.input.Placeholder = "你必须出牌 (如 33344)"
			case m.game.CanBeat():
				m.input.Placeholder = "出牌或 PASS"
			default:
				m.input.Placeholder = "没有能大过上家的牌，输入 PASS"
			}
		}
	}
}

func (m *OnlineModel) processServerMessage(msg ServerMessage) []tea.Cmd {
	var cmds []tea.Cmd
	// Handle server message via injected handler
	if m.serverMessageHandler != nil {
		if cmd := m.serverMessageHandler(m, msg.Msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if m.client.IsConnected() {
		cmds = append(cmds, m.listenForMessages())
	}
	return cmds
}

func (m *OnlineModel) processKeyMsg(msg tea.KeyMsg) (bool, tea.Cmd) {
	// Handle keyboard input via injected handler
	if m.keyHandler != nil {
		handled, keyCmd := m.keyHandler(m, msg)
		if keyCmd != nil {
			return handled, keyCmd
		}
		if handled {
			return true, nil
		}
	}
	return false, nil
}

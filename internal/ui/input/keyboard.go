// Package input handles keyboard input processing.
package input

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
	"github.com/palemoky/fight-the-landlord/internal/ui/view"
)

// sendChatMessage sends a chat message and returns error command if failed
func sendChatMessage(m model.Model, content, scope string) tea.Cmd {
	chatMsg := encoding.MustNewMessage(protocol.MsgChat, protocol.ChatPayload{
		Content: content,
		Scope:   scope,
	})
	if err := m.Client().SendMessage(chatMsg); err != nil {
		m.SetNotification(model.NotifyError, fmt.Sprintf("⚠️ 发送消息失败: %v", err), true)
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return model.ClearSystemNotificationMsg{}
		})
	}
	return nil
}

// HandleKeyPress handles keyboard input and returns whether it was handled.
func HandleKeyPress(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	// Lobby chat handling
	if m.Phase() == model.PhaseLobby {
		chatInput := m.Lobby().ChatInput()

		// "/" key focuses chat input
		if !chatInput.Focused() {
			if msg.String() == "/" {
				chatInput.Focus()
				return true, nil
			}
			goto handleOtherKeys
		}

		// Chat input is focused - handle input
		switch msg.Type {
		case tea.KeyEnter:
			if content := chatInput.Value(); content != "" {
				if cmd := sendChatMessage(m, content, "lobby"); cmd != nil {
					return true, cmd
				}
				chatInput.SetValue("")
			}
			return true, nil
		case tea.KeyEsc:
			chatInput.Blur()
			return true, nil
		default:
			var cmd tea.Cmd
			*chatInput, cmd = chatInput.Update(msg)
			return true, cmd
		}
	}

handleOtherKeys:

	// In-game quick message handling
	if m.Phase() == model.PhaseBidding || m.Phase() == model.PhasePlaying {
		if m.Game().ShowQuickMsgMenu() {
			switch msg.Type {
			case tea.KeyEsc:
				m.Game().SetShowQuickMsgMenu(false)
				return true, nil
			case tea.KeyRunes:
				if msg.String() == "t" || msg.String() == "T" {
					m.Game().SetShowQuickMsgMenu(false)
					return true, nil
				}
				if msg.String() >= "1" && msg.String() <= "8" {
					idx := int(msg.Runes[0] - '1')
					if idx < len(view.QuickMessages) {
						if cmd := sendChatMessage(m, view.QuickMessages[idx], "room"); cmd != nil {
							return true, cmd
						}
						m.Game().SetShowQuickMsgMenu(false)
						return true, nil
					}
				}
			}
			return true, nil
		}

		if msg.String() == "t" || msg.String() == "T" {
			m.Game().SetShowQuickMsgMenu(!m.Game().ShowQuickMsgMenu())
			return true, nil
		}
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return handleEscKey(m)
	case tea.KeyUp:
		m.Lobby().HandleUpKey(m.Phase())
		return false, nil
	case tea.KeyDown:
		m.Lobby().HandleDownKey(m.Phase())
		return false, nil
	case tea.KeyRunes:
		return handleRuneKey(m, msg)
	case tea.KeyEnter:
		cmd := handleEnter(m)
		return false, cmd
	}
	return false, nil
}

func handleEscKey(m model.Model) (bool, tea.Cmd) {
	if m.Game().ShowingHelp() {
		m.Game().SetShowingHelp(false)
		return true, nil
	}

	switch m.Phase() {
	case model.PhaseRoomList, model.PhaseMatching, model.PhaseLeaderboard, model.PhaseStats, model.PhaseRules, model.PhaseGameOver:
		m.EnterLobby()
		return true, nil
	case model.PhaseWaiting:
		_ = m.Client().LeaveRoom()
		m.EnterLobby()
		return true, nil
	case model.PhaseBidding, model.PhasePlaying:
		m.SetNotification(model.NotifyError, "⚠️ 游戏进行中，无法退出！", true)
		return true, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return model.ClearSystemNotificationMsg{}
		})
	}

	m.Client().Close()
	return true, tea.Quit
}

func handleRuneKey(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(msg.Runes) == 0 {
		return false, nil
	}

	if msg.Runes[0] == 'c' || msg.Runes[0] == 'C' {
		if m.Phase() == model.PhaseBidding || m.Phase() == model.PhasePlaying {
			m.Game().SetCardCounterEnabled(!m.Game().CardCounterEnabled())
			return true, nil
		}
	}

	if msg.Runes[0] == 'h' || msg.Runes[0] == 'H' {
		if m.Phase() == model.PhaseBidding || m.Phase() == model.PhasePlaying {
			m.Game().SetShowingHelp(!m.Game().ShowingHelp())
			return true, nil
		}
	}

	return false, nil
}

func handleEnter(m model.Model) tea.Cmd {
	input := strings.TrimSpace(m.Input().Value())
	m.Input().Reset()

	switch m.Phase() {
	case model.PhaseLobby:
		return handleLobbyEnter(m, input)
	case model.PhaseRoomList:
		return handleRoomListEnter(m, input)
	case model.PhaseWaiting:
		return handleWaitingEnter(m, input)
	case model.PhaseBidding:
		return handleBiddingEnter(m, input)
	case model.PhasePlaying:
		return handlePlayingEnter(m, input)
	case model.PhaseGameOver:
		return handleGameOverEnter(m)
	}

	return nil
}

func handleLobbyEnter(m model.Model, input string) tea.Cmd {
	if input == "" {
		input = fmt.Sprintf("%d", m.Lobby().SelectedIndex()+1)
	}

	switch input {
	case "1":
		if m.IsMaintenanceMode() {
			m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return model.ClearSystemNotificationMsg{}
			})
		}
		if m.Client().IsReconnecting() {
			m.SetNotification(model.NotifyError, "⚠️ 正在重连中，请稍后再试", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return model.ClearSystemNotificationMsg{}
			})
		}
		if !m.Client().IsConnected() {
			m.SetNotification(model.NotifyError, "⚠️ 未连接到服务器", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return model.ClearSystemNotificationMsg{}
			})
		}
		m.SetPhase(model.PhaseMatching)
		m.SetMatchingStartTime(time.Now())
		_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgQuickMatch, nil))

	case "2":
		if m.IsMaintenanceMode() {
			m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return model.ClearSystemNotificationMsg{}
			})
		}
		_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgCreateRoom, nil))

	case "3":
		if m.IsMaintenanceMode() {
			m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return model.ClearSystemNotificationMsg{}
			})
		}
		m.SetPhase(model.PhaseRoomList)
		_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgGetRoomList, nil))
		m.Input().Placeholder = "输入房间号或按 ESC 返回"

	case "4":
		m.SetPhase(model.PhaseLeaderboard)
		_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgGetLeaderboard, nil))

	case "5":
		m.SetPhase(model.PhaseStats)
		_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgGetStats, nil))

	case "6":
		m.SetPhase(model.PhaseRules)

	default:
		if m.IsMaintenanceMode() {
			m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return model.ClearSystemNotificationMsg{}
			})
		}
		_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgJoinRoom, protocol.JoinRoomPayload{
			RoomCode: input,
		}))
	}

	return nil
}

func handleRoomListEnter(m model.Model, input string) tea.Cmd {
	if m.IsMaintenanceMode() {
		m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停加入房间", true)
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return model.ClearSystemNotificationMsg{}
		})
	}
	rooms := m.Lobby().AvailableRooms()
	if input == "" {
		if len(rooms) > 0 && m.Lobby().SelectedRoomIdx() < len(rooms) {
			roomCode := rooms[m.Lobby().SelectedRoomIdx()].RoomCode
			_ = m.Client().JoinRoom(roomCode)
		}
	} else {
		_ = m.Client().JoinRoom(input)
	}
	return nil
}

func handleWaitingEnter(m model.Model, input string) tea.Cmd {
	if strings.ToLower(input) == "r" || strings.ToLower(input) == "ready" {
		_ = m.Client().Ready()
	}
	return nil
}

func handleBiddingEnter(m model.Model, input string) tea.Cmd {
	if m.Game().BidTurn() == m.PlayerID() {
		switch strings.ToLower(input) {
		case "y", "yes", "1":
			_ = m.Client().Bid(true)
		case "n", "no", "0":
			_ = m.Client().Bid(false)
		}
	}
	return nil
}

func handlePlayingEnter(m model.Model, input string) tea.Cmd {
	if m.Game().State().CurrentTurn == m.PlayerID() {
		upperInput := strings.ToUpper(input)
		if upperInput == "PASS" || upperInput == "P" {
			_ = m.Client().Pass()
		} else if len(input) > 0 {
			cards, err := card.FindCardsInHand(m.Game().State().Hand, strings.ToUpper(input))
			if err != nil {
				m.Input().Placeholder = err.Error()
				return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return model.ClearInputErrorMsg{}
				})
			}
			_ = m.Client().PlayCards(convert.CardsToInfos(cards))
		}
	}
	return nil
}

func handleGameOverEnter(m model.Model) tea.Cmd {
	m.EnterLobby()
	m.Game().State().Reset()

	_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

	return nil
}

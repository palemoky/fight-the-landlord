// Package input handles keyboard input processing.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
	"github.com/palemoky/fight-the-landlord/internal/ui/view"
)

func clearSystemNotification() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return model.ClearSystemNotificationMsg{}
	})
}

// sendChatMessage sends a chat message and returns error command if failed
func sendChatMessage(m model.Model, content, scope string) tea.Cmd {
	chatMsg := codec.MustNewMessage(protocol.MsgChat, protocol.ChatPayload{
		Content: content,
		Scope:   scope,
	})
	if err := m.Client().SendMessage(chatMsg); err != nil {
		m.SetNotification(model.NotifyError, fmt.Sprintf("⚠️ 发送消息失败: %v", err), true)
		return clearSystemNotification()
	}
	return nil
}

// handleLobbyChatInput handles chat input in the lobby
// Returns (handled, cmd) where handled indicates if the key was processed
func handleLobbyChatInput(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.Phase() != model.PhaseLobby {
		return false, nil
	}

	chatInput := m.Lobby().ChatInput()

	// "/" key focuses chat input
	if !chatInput.Focused() {
		if msg.String() == "/" {
			chatInput.Focus()
			return true, nil
		}
		return false, nil
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

// handleQuickMessageMenu handles the quick message menu in-game
func handleQuickMessageMenu(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.Phase() != model.PhaseBidding && m.Phase() != model.PhasePlaying {
		return false, nil
	}

	// Toggle quick message menu with 'T' key
	if tryToggleQuickMenu(m, msg) {
		return true, nil
	}

	// Handle menu interactions
	if !m.Game().ShowQuickMsgMenu() {
		return false, nil
	}

	return processQuickMenuKey(m, msg)
}

func tryToggleQuickMenu(m model.Model, msg tea.KeyMsg) bool {
	if msg.String() == "t" || msg.String() == "T" {
		if !m.Game().ShowQuickMsgMenu() {
			m.Game().SetShowQuickMsgMenu(true)
			return true
		}
	}
	return false
}

func processQuickMenuKey(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.Game().SetShowQuickMsgMenu(false)
		return true, nil
	case tea.KeyUp, tea.KeyDown:
		return handleQuickMsgScroll(m, msg)
	case tea.KeyEnter:
		return handleQuickMsgEnter(m)
	case tea.KeyBackspace, tea.KeyRunes:
		return handleQuickMsgInput(m, msg)
	}
	return true, nil
}

func handleQuickMsgScroll(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	scroll := m.Game().QuickMsgScroll()
	if msg.Type == tea.KeyUp {
		if scroll > 0 {
			m.Game().SetQuickMsgScroll(scroll - 1)
		}
	} else {
		maxScroll := max(len(view.QuickMessages)-10, 0)
		if scroll < maxScroll {
			m.Game().SetQuickMsgScroll(scroll + 1)
		}
	}
	return true, nil
}

func handleQuickMsgEnter(m model.Model) (bool, tea.Cmd) {
	input := m.Game().QuickMsgInput()
	if input != "" {
		idx := 0
		for _, c := range input {
			idx = idx*10 + int(c-'0')
		}
		idx-- // Convert to 0-indexed
		if idx >= 0 && idx < len(view.QuickMessages) {
			if cmd := sendChatMessage(m, view.QuickMessages[idx], "room"); cmd != nil {
				return true, cmd
			}
			m.Game().SetShowQuickMsgMenu(false)
			return true, nil
		}
	}
	m.Game().ClearQuickMsgInput()
	return true, nil
}

func handleQuickMsgInput(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.Type == tea.KeyBackspace {
		input := m.Game().QuickMsgInput()
		if input != "" {
			m.Game().SetQuickMsgInput(input[:len(input)-1])
		}
		return true, nil
	}

	// KeyRunes
	if msg.String() == "t" || msg.String() == "T" {
		m.Game().SetShowQuickMsgMenu(false)
		return true, nil
	}
	// Accumulate digits for message selection
	if len(msg.Runes) == 1 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
		input := m.Game().QuickMsgInput()
		if len(input) < 2 {
			m.Game().AppendQuickMsgInput(msg.Runes[0])
		}
	}
	return true, nil
}

// HandleKeyPress handles keyboard input and returns whether it was handled.
func HandleKeyPress(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	// Try lobby chat handling first
	if handled, cmd := handleLobbyChatInput(m, msg); handled {
		return true, cmd
	}

	// Try quick message menu handling
	if handled, cmd := handleQuickMessageMenu(m, msg); handled {
		return true, cmd
	}

	// General key handling
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
		return true, clearSystemNotification()
	}

	m.Client().Close()
	return true, tea.Quit
}

func handleRuneKey(m model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(msg.Runes) == 0 {
		return false, nil
	}

	// Handle game toggles (only during bidding/playing)
	if m.Phase() == model.PhaseBidding || m.Phase() == model.PhasePlaying {
		switch msg.Runes[0] {
		case 'c', 'C':
			m.Game().SetCardCounterEnabled(!m.Game().CardCounterEnabled())
			return true, nil
		case 'h', 'H':
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

// checkServerAvailability 检查服务器是否可用于游戏操作
// 返回 true 和错误命令如果服务器不可用，返回 false 和 nil 如果可用
func checkServerAvailability(m model.Model) (blocked bool, cmd tea.Cmd) {
	if blocked, cmd := checkMaintenanceMode(m); blocked {
		return blocked, cmd
	}
	if m.Client().IsReconnecting() {
		m.SetNotification(model.NotifyError, "⚠️ 正在重连中，请稍后再试", true)
		return true, clearSystemNotification()
	}
	if !m.Client().IsConnected() {
		m.SetNotification(model.NotifyError, "⚠️ 未连接到服务器", true)
		return true, clearSystemNotification()
	}
	return false, nil
}

// checkMaintenanceMode 检查服务器是否处于维护模式
// 返回 true 和错误命令如果在维护模式，返回 false 和 nil 如果正常
func checkMaintenanceMode(m model.Model) (blocked bool, cmd tea.Cmd) {
	if m.IsMaintenanceMode() {
		m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
		return true, clearSystemNotification()
	}
	return false, nil
}

func handleLobbyEnter(m model.Model, input string) tea.Cmd {
	if input == "" {
		input = fmt.Sprintf("%d", m.Lobby().SelectedIndex()+1)
	}

	switch input {
	case "1": // 快速匹配
		if blocked, cmd := checkServerAvailability(m); blocked {
			return cmd
		}
		m.SetPhase(model.PhaseMatching)
		m.SetMatchingStartTime(time.Now())
		_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgQuickMatch, nil))

	case "2": // 创建房间
		if blocked, cmd := checkMaintenanceMode(m); blocked {
			return cmd
		}
		_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgCreateRoom, nil))

	case "3": // 房间列表
		if blocked, cmd := checkMaintenanceMode(m); blocked {
			return cmd
		}
		m.SetPhase(model.PhaseRoomList)
		_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetRoomList, nil))
		m.Input().Placeholder = "输入房间号或按 ESC 返回"

	case "4": // 排行榜
		m.SetPhase(model.PhaseLeaderboard)
		_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetLeaderboard, nil))

	case "5": // 统计信息
		m.SetPhase(model.PhaseStats)
		_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetStats, nil))

	case "6": // 游戏规则
		m.SetPhase(model.PhaseRules)

	default: // 加入房间
		if blocked, cmd := checkMaintenanceMode(m); blocked {
			return cmd
		}
		_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgJoinRoom, protocol.JoinRoomPayload{
			RoomCode: input,
		}))
	}

	return nil
}

func handleRoomListEnter(m model.Model, input string) tea.Cmd {
	if m.IsMaintenanceMode() {
		m.SetNotification(model.NotifyError, "⚠️ 服务器维护中，暂停加入房间", true)
		return clearSystemNotification()
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
	if strings.EqualFold(input, "r") || strings.EqualFold(input, "ready") {
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
		} else if input != "" {
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

	_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

	return nil
}

package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// handleKeyPress 处理按键消息，返回是否已处理和命令
func (m *OnlineModel) handleKeyPress(msg tea.KeyMsg) (bool, tea.Cmd) {
	// 全局 Chat Chat Focus 切换 (大厅)
	if m.phase == PhaseLobby {
		if m.lobby.chatInput.Focused() {
			switch msg.Type {
			case tea.KeyEnter:
				// 发送消息
				content := m.lobby.chatInput.Value()
				if content != "" {
					chatMsg := protocol.MustNewMessage(protocol.MsgChat, protocol.ChatPayload{
						Content: content,
						Scope:   "lobby",
					})
					if err := m.client.SendMessage(chatMsg); err != nil {
						m.setNotification(NotifyError, fmt.Sprintf("⚠️ 发送消息失败: %v", err), true)
						return true, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
							return ClearSystemNotificationMsg{}
						})
					}
					m.lobby.chatInput.SetValue("")
				}
				return true, nil
			case tea.KeyEsc:
				m.lobby.chatInput.Blur()
				return true, nil
			default:
				var cmd tea.Cmd
				m.lobby.chatInput, cmd = m.lobby.chatInput.Update(msg)
				return true, cmd
			}
		} else if msg.String() == "/" {
			m.lobby.chatInput.Focus()
			return true, nil
		}
	}

	// 游戏内 Quick Message (only during actual gameplay, not waiting)
	isInGame := m.phase == PhaseBidding || m.phase == PhasePlaying
	if isInGame {
		// 处理快捷消息菜单
		if m.game.showQuickMsgMenu {
			switch msg.Type {
			case tea.KeyEsc:
				m.game.showQuickMsgMenu = false
				return true, nil
			case tea.KeyRunes:
				// T 键关闭菜单
				if msg.String() == "t" || msg.String() == "T" {
					m.game.showQuickMsgMenu = false
					return true, nil
				}
				// 数字键选择 1-8
				if msg.String() >= "1" && msg.String() <= "8" {
					idx := int(msg.Runes[0] - '1')
					if idx < len(quickMessages) {
						content := quickMessages[idx]
						chatMsg := protocol.MustNewMessage(protocol.MsgChat, protocol.ChatPayload{
							Content: content,
							Scope:   "room",
						})
						if err := m.client.SendMessage(chatMsg); err != nil {
							m.setNotification(NotifyError, fmt.Sprintf("⚠️ 发送消息失败: %v", err), true)
							return true, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
								return ClearSystemNotificationMsg{}
							})
						}
						m.game.showQuickMsgMenu = false
						return true, nil
					}
				}
			}
			return true, nil // 吞掉其他按键，模态
		}

		// T 键切换快捷消息菜单
		if msg.String() == "t" || msg.String() == "T" {
			m.game.showQuickMsgMenu = !m.game.showQuickMsgMenu
			return true, nil
		}
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m.handleEscKey()
	case tea.KeyUp:
		m.handleUpKey()
		return false, nil
	case tea.KeyDown:
		m.handleDownKey()
		return false, nil
	case tea.KeyRunes:
		return m.handleRuneKey(msg)
	case tea.KeyEnter:
		cmd := m.handleEnter()
		return false, cmd
	}
	return false, nil
}

// handleEscKey 处理 ESC 键
func (m *OnlineModel) handleEscKey() (bool, tea.Cmd) {
	// 如果游戏中正在显示帮助，先关闭帮助
	if m.game.showingHelp {
		m.game.showingHelp = false
		return true, nil
	}
	// 从特定页面返回大厅（直接返回，无需额外操作）
	if m.phase == PhaseRoomList || m.phase == PhaseMatching || m.phase == PhaseLeaderboard || m.phase == PhaseStats || m.phase == PhaseRules || m.phase == PhaseGameOver {
		m.enterLobby()
		return true, nil
	}
	// 从等待房间返回大厅（需要先通知服务器离开房间）
	if m.phase == PhaseWaiting {
		_ = m.client.LeaveRoom()
		m.enterLobby()
		return true, nil
	}
	// 在游戏中（叫地主、出牌）时，ESC 不退出游戏，避免误操作
	if m.phase == PhaseBidding || m.phase == PhasePlaying {
		// 显示提示信息，3秒后自动消失
		m.setNotification(NotifyError, "⚠️ 游戏进行中，无法退出！", true)
		return true, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearSystemNotificationMsg{}
		})
	}
	// 其他情况（大厅、游戏结束等）可以退出
	m.client.Close()
	return true, tea.Quit
}

// handleUpKey 处理上箭头键
func (m *OnlineModel) handleUpKey() {
	m.lobby.handleUpKey(m.phase)
}

// handleDownKey 处理下箭头键
func (m *OnlineModel) handleDownKey() {
	m.lobby.handleDownKey(m.phase)
}

// handleRuneKey 处理字符键（C/H 等）
func (m *OnlineModel) handleRuneKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(msg.Runes) == 0 {
		return false, nil
	}

	// C 键切换记牌器
	if msg.Runes[0] == 'c' || msg.Runes[0] == 'C' {
		if m.phase == PhaseBidding || m.phase == PhasePlaying {
			m.game.cardCounterEnabled = !m.game.cardCounterEnabled
			// 直接返回，不让 textinput 处理这个按键
			return true, nil
		}
	}

	// H 键切查看帮助（R 会与大王键冲突）
	if msg.Runes[0] == 'h' || msg.Runes[0] == 'H' {
		if m.phase == PhaseBidding || m.phase == PhasePlaying {
			m.game.showingHelp = !m.game.showingHelp
			// 直接返回，不让 textinput 处理这个按键
			return true, nil
		}
	}

	return false, nil
}

// handleTimeout 处理超时消息
func (m *OnlineModel) handleTimeout() {
	if m.phase == PhaseBidding && m.game.bidTurn == m.playerID {
		_ = m.client.Bid(false) // 自动不叫
	} else if m.phase == PhasePlaying && m.game.state.CurrentTurn == m.playerID {
		if m.game.mustPlay && len(m.game.state.Hand) > 0 {
			// 自动出最小的牌
			minCard := m.game.state.Hand[len(m.game.state.Hand)-1]
			_ = m.client.PlayCards([]protocol.CardInfo{protocol.CardToInfo(minCard)})
		} else {
			_ = m.client.Pass()
		}
	}
}

// handleEnter 处理回车键
func (m *OnlineModel) handleEnter() tea.Cmd {
	input := strings.TrimSpace(m.input.Value())
	m.input.Reset()
	m.error = ""

	switch m.phase {
	case PhaseLobby:
		return m.handleLobbyEnter(input)
	case PhaseRoomList:
		return m.handleRoomListEnter(input)
	case PhaseWaiting:
		return m.handleWaitingEnter(input)
	case PhaseBidding:
		return m.handleBiddingEnter(input)
	case PhasePlaying:
		return m.handlePlayingEnter(input)
	case PhaseGameOver:
		return m.handleGameOverEnter()
	}

	return nil
}

// handleLobbyEnter 处理大厅界面的回车
func (m *OnlineModel) handleLobbyEnter(input string) tea.Cmd {
	// 如果没有输入，使用当前选中的菜单项
	if input == "" {
		input = fmt.Sprintf("%d", m.lobby.selectedIndex+1)
	}

	switch input {
	case "1":
		// 快速匹配 - 先检查状态
		// 检查是否在维护模式
		if m.maintenanceMode {
			m.setNotification(NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return ClearSystemNotificationMsg{}
			})
		}

		// 检查是否正在重连
		if m.client.IsReconnecting() {
			m.setNotification(NotifyError, "⚠️ 正在重连中，请稍后再试", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return ClearSystemNotificationMsg{}
			})
		}

		// 检查连接状态
		if !m.client.IsConnected() {
			m.setNotification(NotifyError, "⚠️ 未连接到服务器", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return ClearSystemNotificationMsg{}
			})
		}

		// 状态检查通过，进入匹配
		m.phase = PhaseMatching
		m.matchingStartTime = time.Now()
		_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgQuickMatch, nil))

	case "2":
		// 创建房间
		if m.maintenanceMode {
			m.setNotification(NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return ClearSystemNotificationMsg{}
			})
		}
		_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgCreateRoom, nil))

	case "3":
		// 加入房间 - 显示房间列表
		if m.maintenanceMode {
			m.setNotification(NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return ClearSystemNotificationMsg{}
			})
		}
		m.phase = PhaseRoomList
		_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgGetRoomList, nil))
		m.input.Placeholder = "输入房间号或按 ESC 返回"

	case "4":
		// 排行榜
		m.phase = PhaseLeaderboard
		_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgGetLeaderboard, nil))

	case "5":
		// 我的战绩
		m.phase = PhaseStats
		_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgGetStats, nil))

	case "6":
		// 游戏规则
		m.phase = PhaseRules

	default:
		// 尝试作为房间号处理
		if m.maintenanceMode {
			m.setNotification(NotifyError, "⚠️ 服务器维护中，暂停接受新连接", true)
			return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return ClearSystemNotificationMsg{}
			})
		}
		_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgJoinRoom, protocol.JoinRoomPayload{
			RoomCode: input,
		}))
	}

	return nil
}

// handleRoomListEnter 处理房间列表界面的回车
func (m *OnlineModel) handleRoomListEnter(input string) tea.Cmd {
	if m.maintenanceMode {
		m.setNotification(NotifyError, "⚠️ 服务器维护中，暂停加入房间", true)
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearSystemNotificationMsg{}
		})
	}
	if input == "" {
		// 没有输入，加入选中的房间
		if len(m.lobby.availableRooms) > 0 && m.lobby.selectedRoomIdx < len(m.lobby.availableRooms) {
			roomCode := m.lobby.availableRooms[m.lobby.selectedRoomIdx].RoomCode
			_ = m.client.JoinRoom(roomCode)
		}
	} else {
		// 有输入，直接加入输入的房间号
		_ = m.client.JoinRoom(input)
	}
	return nil
}

// handleWaitingEnter 处理等待房间的回车
func (m *OnlineModel) handleWaitingEnter(input string) tea.Cmd {
	if strings.ToLower(input) == "r" || strings.ToLower(input) == "ready" {
		_ = m.client.Ready()
	}
	return nil
}

// handleBiddingEnter 处理叫地主阶段的回车
func (m *OnlineModel) handleBiddingEnter(input string) tea.Cmd {
	if m.game.bidTurn == m.playerID {
		switch strings.ToLower(input) {
		case "y", "yes", "1":
			_ = m.client.Bid(true)
		case "n", "no", "0":
			_ = m.client.Bid(false)
		}
	}
	return nil
}

// handlePlayingEnter 处理出牌阶段的回车
func (m *OnlineModel) handlePlayingEnter(input string) tea.Cmd {
	if m.game.state.CurrentTurn == m.playerID {
		upperInput := strings.ToUpper(input)
		if upperInput == "PASS" || upperInput == "P" {
			_ = m.client.Pass()
		} else if len(input) > 0 {
			// 解析出牌
			cards, err := m.parseCardsInput(input)
			if err != nil {
				// 显示在 placeholder，3秒后清除
				m.input.Placeholder = err.Error()
				return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return ClearInputErrorMsg{}
				})
			} else {
				_ = m.client.PlayCards(protocol.CardsToInfos(cards))
			}
		}
	}
	return nil
}

// handleGameOverEnter 处理游戏结束的回车
func (m *OnlineModel) handleGameOverEnter() tea.Cmd {
	m.enterLobby()
	m.resetGameState()

	// 返回大厅时查询维护状态和在线人数
	_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

	return nil
}

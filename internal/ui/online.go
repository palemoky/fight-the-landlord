package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/client"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/sound"
)

// 游戏阶段
type GamePhase int

const (
	PhaseConnecting GamePhase = iota
	PhaseReconnecting
	PhaseLobby
	PhaseRoomList
	PhaseMatching
	PhaseWaiting
	PhaseBidding
	PhasePlaying
	PhaseGameOver
	PhaseLeaderboard
	PhaseStats
	PhaseRules
)

// ServerMessage 服务器消息（用于 tea.Msg）
type ServerMessage struct {
	Msg *protocol.Message
}

// ConnectedMsg 连接成功消息
type ConnectedMsg struct{}

// ConnectionErrorMsg 连接错误消息
type ConnectionErrorMsg struct {
	Err error
}

// ReconnectingMsg 正在重连消息
type ReconnectingMsg struct {
	Attempt  int
	MaxTries int
}

// ReconnectSuccessMsg 重连成功消息
type ReconnectSuccessMsg struct{}

// ClearReconnectMsg 清除重连消息
type ClearReconnectMsg struct{}

// ClearErrorMsg 清除错误消息
type ClearErrorMsg struct{}

// OnlineModel 联网模式的 model
type OnlineModel struct {
	client *client.Client
	phase  GamePhase
	error  string

	// 玩家信息
	playerID   string
	playerName string

	matchingStartTime time.Time // 匹配开始时间

	// 网络状态
	latency int64 // 延迟（毫秒）

	// 重连状态
	reconnecting      bool         // 是否正在重连
	reconnectAttempt  int          // 当前重连尝试次数
	reconnectMaxTries int          // 最大重连次数
	reconnectSuccess  bool         // 重连是否成功
	reconnectMessage  string       // 重连消息
	reconnectChan     chan tea.Msg // 重连消息通道（可发送多种消息类型）

	// Sub-models
	lobby *LobbyModel
	game  *GameModel

	// Audio
	soundManager *sound.SoundManager

	// UI 组件
	input  *textinput.Model
	timer  timer.Model
	width  int
	height int
}

// NewOnlineModel 创建联网模式 model
func NewOnlineModel(serverURL string) *OnlineModel {
	ti := textinput.New()
	ti.Placeholder = "输入房间号..."
	ti.CharLimit = 10
	ti.Width = 20
	ti.Focus()

	c := client.NewClient(serverURL)
	reconnectChan := make(chan tea.Msg, 10)

	m := &OnlineModel{
		client:            c,
		phase:             PhaseConnecting,
		input:             &ti,
		reconnectMaxTries: 5, // 最大重连次数
		reconnectChan:     reconnectChan,
		lobby:             NewLobbyModel(c, &ti), // Pass pointer to shared input
		game:              NewGameModel(c, &ti),  // Pass pointer to shared input
		soundManager:      sound.NewSoundManager(),
	}

	// 设置重连回调 - 通过 channel 发送消息到 Bubble Tea
	c.OnReconnecting = func(attempt, maxTries int) {
		select {
		case reconnectChan <- ReconnectingMsg{Attempt: attempt, MaxTries: maxTries}:
		default:
		}
	}

	// 设置重连成功回调
	c.OnReconnect = func() {
		select {
		case reconnectChan <- ReconnectSuccessMsg{}:
		default:
		}
	}

	return m
}

func (m *OnlineModel) Init() tea.Cmd {
	// Initialize sound
	go func() {
		_ = m.soundManager.Init()
	}()

	return tea.Batch(
		m.connectToServer(),
		textinput.Blink,
		m.listenForReconnect(),
	)
}

// listenForReconnect 监听重连消息
func (m *OnlineModel) listenForReconnect() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.reconnectChan
		return msg
	}
}

// connectToServer 连接服务器
func (m *OnlineModel) connectToServer() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.Connect(); err != nil {
			return ConnectionErrorMsg{Err: err}
		}
		return ConnectedMsg{}
	}
}

// listenForMessages 监听服务器消息
func (m *OnlineModel) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		msg, err := m.client.Receive()
		if err != nil {
			return ConnectionErrorMsg{Err: err}
		}
		return ServerMessage{Msg: msg}
	}
}

func (m *OnlineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.lobby.width = msg.Width
		m.lobby.height = msg.Height
		m.game.width = msg.Width
		m.game.height = msg.Height

	case tea.KeyMsg:
		// 提取按键处理到独立方法
		handled, returnCmd := m.handleKeyPress(msg)
		if handled {
			return m, returnCmd
		}

	case ConnectedMsg:
		m.phase = PhaseLobby
		m.playerID = m.client.PlayerID
		m.playerName = m.client.PlayerName
		// 启动心跳
		m.client.StartHeartbeat()
		// 开始监听消息
		cmds = append(cmds, m.listenForMessages())

	case ConnectionErrorMsg:
		m.error = fmt.Sprintf("无法连接到服务器: %v\n\n按 ESC 退出", msg.Err)
		// 保持在连接阶段，不显示大厅菜单
		m.phase = PhaseConnecting

	case ReconnectingMsg:
		m.reconnecting = true
		m.reconnectAttempt = msg.Attempt
		m.reconnectMaxTries = msg.MaxTries
		m.reconnectSuccess = false
		m.reconnectMessage = fmt.Sprintf("🔄 正在重连 (%d/%d)...", msg.Attempt, msg.MaxTries)
		// 继续监听重连消息
		cmds = append(cmds, m.listenForReconnect())

	case ReconnectSuccessMsg:
		m.reconnecting = false
		m.reconnectSuccess = true
		m.reconnectMessage = "✅ 重连成功！"
		// 3秒后清除消息
		cmds = append(cmds, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearReconnectMsg{}
		}))
		// 继续监听重连消息（为未来的重连做准备）
		cmds = append(cmds, m.listenForReconnect())
		// 重新开始监听服务器消息（因为重连后 receive channel 被重置了）
		if m.client.IsConnected() {
			cmds = append(cmds, m.listenForMessages())
		}

	case ClearReconnectMsg:
		m.reconnectSuccess = false
		m.reconnectMessage = ""

	case ClearErrorMsg:
		m.error = ""

	case ServerMessage:
		cmd = m.handleServerMessage(msg.Msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// 继续监听
		if m.client.IsConnected() {
			cmds = append(cmds, m.listenForMessages())
		}

	case timer.TimeoutMsg:
		m.handleTimeout()

	case timer.TickMsg:
		// 检查是否需要播放提示音
		if m.shouldPlayBell() {
			m.game.bellPlayed = true
			cmds = append(cmds, m.playBell())
		}
	}

	m.timer, cmd = m.timer.Update(msg)
	cmds = append(cmds, cmd)

	// Update the input model (dereference the pointer)
	newInput, cmd := m.input.Update(msg)
	*m.input = newInput // Update the value at the pointer address
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

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
						m.error = fmt.Sprintf("发送消息失败: %v", err)
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

	// 游戏内 Quick Message (no chat input, only quick messages)
	isInGame := m.phase == PhaseWaiting || m.phase == PhaseBidding || m.phase == PhasePlaying
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
							m.error = fmt.Sprintf("发送消息失败: %v", err)
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
	// 从特定页面返回大厅
	if m.phase == PhaseRoomList || m.phase == PhaseMatching || m.phase == PhaseLeaderboard || m.phase == PhaseStats || m.phase == PhaseRules {
		m.phase = PhaseLobby
		m.error = ""
		m.input.Reset()
		m.input.Placeholder = "输入选项 (1-6) 或房间号"
		m.input.Focus()
		return true, nil
	}
	// 在游戏中（等待、叫地主、出牌）时，ESC 不退出游戏，避免误操作
	if m.phase == PhaseWaiting || m.phase == PhaseBidding || m.phase == PhasePlaying {
		// 显示提示信息，3秒后自动消失
		m.error = "游戏进行中，无法退出！"
		return true, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearErrorMsg{}
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
	} else if m.phase == PhasePlaying && m.game.currentTurn == m.playerID {
		if m.game.mustPlay && len(m.game.hand) > 0 {
			// 自动出最小的牌
			minCard := m.game.hand[len(m.game.hand)-1]
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
		// 大厅界面：1=快速匹配, 2=创建房间, 3=加入房间, 4=排行榜, 5=我的战绩, 6=游戏规则
		// 如果输入为空，使用选中的菜单项
		if input == "" {
			input = fmt.Sprintf("%d", m.lobby.selectedIndex+1)
		}

		switch input {
		case "1":
			m.phase = PhaseMatching
			m.matchingStartTime = time.Now()
			_ = m.client.QuickMatch()
		case "2":
			_ = m.client.CreateRoom()
		case "3":
			// 请求房间列表
			m.phase = PhaseRoomList
			m.lobby.selectedRoomIdx = 0
			m.input.Placeholder = "或直接输入房间号..."
			m.input.Focus()
			_ = m.client.GetRoomList()
		case "4":
			m.phase = PhaseLeaderboard
			_ = m.client.GetLeaderboard("total", 0, 10)
		case "5":
			m.phase = PhaseStats
			_ = m.client.GetStats()
		case "6":
			m.phase = PhaseRules
		default:
			// 可能是房间号
			if len(input) > 0 {
				_ = m.client.JoinRoom(input)
			}
		}

	case PhaseRoomList:
		// 房间列表界面
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

	case PhaseWaiting:
		// 等待房间：输入 r 准备
		if strings.ToLower(input) == "r" || strings.ToLower(input) == "ready" {
			_ = m.client.Ready()
		}

	case PhaseBidding:
		// 叫地主：y=叫, n=不叫
		if m.game.bidTurn == m.playerID {
			switch strings.ToLower(input) {
			case "y", "yes", "1":
				_ = m.client.Bid(true)
			case "n", "no", "0":
				_ = m.client.Bid(false)
			}
		}

	case PhasePlaying:
		// 出牌
		if m.game.currentTurn == m.playerID {
			upperInput := strings.ToUpper(input)
			if upperInput == "PASS" || upperInput == "P" {
				_ = m.client.Pass()
			} else if len(input) > 0 {
				// 解析出牌
				cards, err := m.parseCardsInput(input)
				if err != nil {
					m.error = err.Error()
				} else {
					_ = m.client.PlayCards(protocol.CardsToInfos(cards))
				}
			}
		}

	case PhaseGameOver:
		// 游戏结束：输入任意键返回大厅
		m.phase = PhaseLobby
		m.input.Placeholder = "输入选项 (1-5) 或房间号"
		m.input.Focus()
		m.resetGameState()
	}

	return nil
}

func (m *OnlineModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.phase {
	case PhaseConnecting:
		content = m.connectingView()
	case PhaseLobby:
		content = m.lobby.lobbyView(m)
	case PhaseRoomList:
		content = m.lobby.roomListView(m)
	case PhaseMatching:
		content = m.matchingView()
	case PhaseWaiting:
		content = m.game.waitingView(m)
	case PhaseBidding, PhasePlaying:
		content = m.game.gameView(m)
	case PhaseGameOver:
		content = m.game.gameOverView()
	case PhaseLeaderboard:
		content = m.lobby.leaderboardView()
	case PhaseStats:
		content = m.lobby.statsView()
	case PhaseRules:
		content = m.game.rulesView()
	}

	return docStyle.Render(content)
}

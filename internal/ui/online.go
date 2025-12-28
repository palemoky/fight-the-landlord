package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

// NotificationType 通知类型
type NotificationType int

const (
	NotifyError            NotificationType = iota // 错误信息（临时）
	NotifyRateLimit                                // 限频提示（临时）
	NotifyReconnecting                             // 重连中（持久）
	NotifyReconnectSuccess                         // 重连成功（临时）
	NotifyMaintenance                              // 维护通知（持久）
	NotifyOnlineCount                              // 在线人数（持久）
)

// SystemNotification 系统通知
type SystemNotification struct {
	Message   string
	Type      NotificationType
	Temporary bool // 是否为临时通知（3秒后自动消失）
}

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

// ClearInputErrorMsg 清除输入框错误消息
type ClearInputErrorMsg struct{}

// ClearSystemNotificationMsg 清除系统通知消息
type ClearSystemNotificationMsg struct{}

// OnlineModel 联网模式的 model
type OnlineModel struct {
	client *client.Client
	phase  GamePhase
	error  string // 保留用于游戏阶段的输入框错误显示

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
	reconnectChan     chan tea.Msg // 重连消息通道（可发送多种消息类型）

	// 维护模式
	maintenanceMode bool // 服务器是否在维护模式

	// 系统通知（统一管理所有通知）
	notifications map[NotificationType]*SystemNotification // 按类型存储的通知

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
		notifications:     make(map[NotificationType]*SystemNotification),
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

// setNotification 设置通知
func (m *OnlineModel) setNotification(notifyType NotificationType, message string, temporary bool) {
	m.notifications[notifyType] = &SystemNotification{
		Message:   message,
		Type:      notifyType,
		Temporary: temporary,
	}
}

// clearNotification 清除指定类型的通知
func (m *OnlineModel) clearNotification(notifyType NotificationType) {
	delete(m.notifications, notifyType)
}

// getCurrentNotification 根据优先级获取当前应显示的通知
// 优先级: 错误 > 限频 > 重连中 > 重连成功 > 维护 > 在线人数
func (m *OnlineModel) getCurrentNotification() *SystemNotification {
	// 按优先级顺序检查
	priorityOrder := []NotificationType{
		NotifyError,
		NotifyRateLimit,
		NotifyReconnecting,
		NotifyReconnectSuccess,
		NotifyMaintenance,
		NotifyOnlineCount,
	}

	for _, notifyType := range priorityOrder {
		if notification, exists := m.notifications[notifyType]; exists {
			return notification
		}
	}

	return nil
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
		m.enterLobby()
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
		// 设置重连中通知（持久显示）
		m.setNotification(NotifyReconnecting, fmt.Sprintf("🔄 正在重连 (%d/%d)...", msg.Attempt, msg.MaxTries), false)
		// 继续监听重连消息
		cmds = append(cmds, m.listenForReconnect())

	case ReconnectSuccessMsg:
		m.reconnecting = false
		// 清除重连中通知
		m.clearNotification(NotifyReconnecting)
		// 设置重连成功通知（临时显示，3秒后消失）
		m.setNotification(NotifyReconnectSuccess, "✅ 重连成功！", true)
		// 3秒后清除重连成功消息并请求在线人数
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
		// 清除重连成功通知
		m.clearNotification(NotifyReconnectSuccess)
		// 如果在大厅，请求在线人数和维护状态
		if m.phase == PhaseLobby {
			_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgGetOnlineCount, nil))
			_ = m.client.SendMessage(protocol.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))
		}

	case ClearErrorMsg:
		m.error = ""

	case ClearInputErrorMsg:
		// 恢复游戏阶段的默认 placeholder
		if m.phase == PhaseBidding && m.game.bidTurn == m.playerID {
			m.input.Placeholder = "叫地主? (Y/N)"
		} else if m.phase == PhasePlaying && m.game.currentTurn == m.playerID {
			switch {
			case m.game.mustPlay:
				m.input.Placeholder = "你必须出牌 (如 33344)"
			case m.game.canBeat:
				m.input.Placeholder = "出牌或 PASS"
			default:
				m.input.Placeholder = "没有能大过上家的牌，输入 PASS"
			}
		}

	case ClearSystemNotificationMsg:
		// 清除临时通知（错误、限频等）
		m.clearNotification(NotifyError)
		m.clearNotification(NotifyRateLimit)

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

// enterLobby enters the lobby phase
func (m *OnlineModel) enterLobby() {
	m.phase = PhaseLobby
	m.error = ""
	m.input.Reset()
	m.input.Placeholder = "输入选项 (1-6) 或房间号"
	m.input.Focus()
	// Note: Online count is requested in handleMsgConnected, no need to request again here
}

// connectingView 显示连接中状态
func (m *OnlineModel) connectingView() string {
	var sb string
	if m.error != "" {
		sb = errorStyle.Render(m.error)
	} else {
		sb = "正在连接服务器..."
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb)
}

// matchingView 显示匹配中状态
func (m *OnlineModel) matchingView() string {
	elapsed := time.Since(m.matchingStartTime).Seconds()
	msg := fmt.Sprintf("🔍 正在匹配玩家...\n\n已等待: %.0f 秒\n\n按 ESC 取消", elapsed)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}

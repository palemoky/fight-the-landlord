package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/client"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// 游戏阶段
type GamePhase int

const (
	PhaseConnecting GamePhase = iota
	PhaseReconnecting
	PhaseLobby
	PhaseMatching
	PhaseWaiting
	PhaseBidding
	PhasePlaying
	PhaseGameOver
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

// OnlineModel 联网模式的 model
type OnlineModel struct {
	client *client.Client
	phase  GamePhase
	error  string

	// 玩家信息
	playerID   string
	playerName string

	// 房间信息
	roomCode string
	players  []protocol.PlayerInfo

	// 游戏状态
	hand           []card.Card
	landlordCards  []card.Card
	currentTurn    string // 当前回合玩家 ID
	lastPlayedBy   string
	lastPlayedName string
	lastPlayed     []card.Card
	lastHandType   string
	mustPlay       bool
	canBeat        bool
	isLandlord     bool

	// 叫地主
	bidTurn string

	// 游戏结束
	winner           string
	winnerIsLandlord bool

	// 网络状态
	latency int64 // 延迟（毫秒）

	// 提醒状态
	bellPlayed bool // 是否已播放提示音

	// 排行榜
	myStats     *protocol.StatsResultPayload
	leaderboard []protocol.LeaderboardEntry

	// UI 组件
	input  textinput.Model
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

	c := client.NewClient(serverURL)

	return &OnlineModel{
		client: c,
		phase:  PhaseConnecting,
		input:  ti,
	}
}

func (m *OnlineModel) Init() tea.Cmd {
	return tea.Batch(
		m.connectToServer(),
		textinput.Blink,
	)
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

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.client.Close()
			return m, tea.Quit
		case tea.KeyEnter:
			cmd = m.handleEnter()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
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
		m.error = fmt.Sprintf("连接错误: %v", msg.Err)
		m.phase = PhaseLobby

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
		// 超时处理
		if m.phase == PhaseBidding && m.bidTurn == m.playerID {
			_ = m.client.Bid(false) // 自动不叫
		} else if m.phase == PhasePlaying && m.currentTurn == m.playerID {
			if m.mustPlay && len(m.hand) > 0 {
				// 自动出最小的牌
				minCard := m.hand[len(m.hand)-1]
				_ = m.client.PlayCards([]protocol.CardInfo{protocol.CardToInfo(minCard)})
			} else {
				_ = m.client.Pass()
			}
		}

	case timer.TickMsg:
		// 检查是否需要播放提示音
		if m.shouldPlayBell() {
			m.bellPlayed = true
			cmds = append(cmds, m.playBell())
		}
	}

	m.timer, cmd = m.timer.Update(msg)
	cmds = append(cmds, cmd)

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleEnter 处理回车键
func (m *OnlineModel) handleEnter() tea.Cmd {
	input := strings.TrimSpace(m.input.Value())
	m.input.Reset()
	m.error = ""

	switch m.phase {
	case PhaseLobby:
		// 大厅界面：1=创建房间, 2=加入房间, 3=快速匹配, 4=排行榜, 5=我的战绩
		switch input {
		case "1":
			_ = m.client.CreateRoom()
		case "2":
			m.input.Placeholder = "请输入房间号..."
			m.input.Focus()
		case "3":
			m.phase = PhaseMatching
			_ = m.client.QuickMatch()
		case "4":
			_ = m.client.GetLeaderboard("total", 0, 10)
		case "5":
			_ = m.client.GetStats()
		default:
			// 可能是房间号
			if len(input) > 0 {
				_ = m.client.JoinRoom(input)
			}
		}

	case PhaseWaiting:
		// 等待房间：输入 r 准备
		if strings.ToLower(input) == "r" || strings.ToLower(input) == "ready" {
			_ = m.client.Ready()
		}

	case PhaseBidding:
		// 叫地主：y=叫, n=不叫
		if m.bidTurn == m.playerID {
			switch strings.ToLower(input) {
			case "y", "yes", "1":
				_ = m.client.Bid(true)
			case "n", "no", "0":
				_ = m.client.Bid(false)
			}
		}

	case PhasePlaying:
		// 出牌
		if m.currentTurn == m.playerID {
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
		content = m.lobbyView()
	case PhaseMatching:
		content = m.matchingView()
	case PhaseWaiting:
		content = m.waitingView()
	case PhaseBidding, PhasePlaying:
		content = m.gameView()
	case PhaseGameOver:
		content = m.gameOverView()
	}

	return docStyle.Render(content)
}

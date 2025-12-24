package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/palemoky/fight-the-landlord-go/internal/card"
	"github.com/palemoky/fight-the-landlord-go/internal/network/client"
	"github.com/palemoky/fight-the-landlord-go/internal/network/protocol"
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
			m.client.Bid(false) // 自动不叫
		} else if m.phase == PhasePlaying && m.currentTurn == m.playerID {
			if m.mustPlay && len(m.hand) > 0 {
				// 自动出最小的牌
				minCard := m.hand[len(m.hand)-1]
				m.client.PlayCards([]protocol.CardInfo{protocol.CardToInfo(minCard)})
			} else {
				m.client.Pass()
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
		// 大厅界面：1=创建房间, 2=加入房间, 3=快速匹配
		switch input {
		case "1":
			m.client.CreateRoom()
		case "2":
			m.input.Placeholder = "请输入房间号..."
			m.input.Focus()
		case "3":
			m.phase = PhaseMatching
			m.client.QuickMatch()
		default:
			// 可能是房间号
			if len(input) > 0 {
				m.client.JoinRoom(input)
			}
		}

	case PhaseWaiting:
		// 等待房间：输入 r 准备
		if strings.ToLower(input) == "r" || strings.ToLower(input) == "ready" {
			m.client.Ready()
		}

	case PhaseBidding:
		// 叫地主：y=叫, n=不叫
		if m.bidTurn == m.playerID {
			switch strings.ToLower(input) {
			case "y", "yes", "1":
				m.client.Bid(true)
			case "n", "no", "0":
				m.client.Bid(false)
			}
		}

	case PhasePlaying:
		// 出牌
		if m.currentTurn == m.playerID {
			upperInput := strings.ToUpper(input)
			if upperInput == "PASS" || upperInput == "P" {
				m.client.Pass()
			} else if len(input) > 0 {
				// 解析出牌
				cards, err := m.parseCardsInput(input)
				if err != nil {
					m.error = err.Error()
				} else {
					m.client.PlayCards(protocol.CardsToInfos(cards))
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

// handleServerMessage 处理服务器消息
func (m *OnlineModel) handleServerMessage(msg *protocol.Message) tea.Cmd {
	switch msg.Type {
	case protocol.MsgConnected:
		var payload protocol.ConnectedPayload
		json.Unmarshal(msg.Payload, &payload)
		m.playerID = payload.PlayerID
		m.playerName = payload.PlayerName

	case protocol.MsgRoomCreated:
		var payload protocol.RoomCreatedPayload
		json.Unmarshal(msg.Payload, &payload)
		m.roomCode = payload.RoomCode
		m.players = []protocol.PlayerInfo{payload.Player}
		m.phase = PhaseWaiting
		m.input.Placeholder = "输入 R 准备"

	case protocol.MsgRoomJoined:
		var payload protocol.RoomJoinedPayload
		json.Unmarshal(msg.Payload, &payload)
		m.roomCode = payload.RoomCode
		m.players = payload.Players
		m.phase = PhaseWaiting
		m.input.Placeholder = "输入 R 准备"

	case protocol.MsgPlayerJoined:
		var payload protocol.PlayerJoinedPayload
		json.Unmarshal(msg.Payload, &payload)
		m.players = append(m.players, payload.Player)

	case protocol.MsgPlayerLeft:
		var payload protocol.PlayerLeftPayload
		json.Unmarshal(msg.Payload, &payload)
		for i, p := range m.players {
			if p.ID == payload.PlayerID {
				m.players = append(m.players[:i], m.players[i+1:]...)
				break
			}
		}

	case protocol.MsgPlayerReady:
		var payload protocol.PlayerReadyPayload
		json.Unmarshal(msg.Payload, &payload)
		for i, p := range m.players {
			if p.ID == payload.PlayerID {
				m.players[i].Ready = payload.Ready
				break
			}
		}

	case protocol.MsgGameStart:
		var payload protocol.GameStartPayload
		json.Unmarshal(msg.Payload, &payload)
		m.players = payload.Players

	case protocol.MsgDealCards:
		var payload protocol.DealCardsPayload
		json.Unmarshal(msg.Payload, &payload)
		m.hand = protocol.InfosToCards(payload.Cards)
		m.sortHand()
		if len(payload.LandlordCards) > 0 && payload.LandlordCards[0].Rank > 0 {
			m.landlordCards = protocol.InfosToCards(payload.LandlordCards)
		}

	case protocol.MsgBidTurn:
		var payload protocol.BidTurnPayload
		json.Unmarshal(msg.Payload, &payload)
		m.phase = PhaseBidding
		m.bidTurn = payload.PlayerID
		m.resetBell() // 重置提示音状态
		if payload.PlayerID == m.playerID {
			m.input.Placeholder = "叫地主? (Y/N)"
			m.input.Focus()
		}
		m.timer = timer.NewWithInterval(time.Duration(payload.Timeout)*time.Second, time.Second)
		return m.timer.Start()

	case protocol.MsgBidResult:
		// 可以显示叫地主结果

	case protocol.MsgLandlord:
		var payload protocol.LandlordPayload
		json.Unmarshal(msg.Payload, &payload)
		m.landlordCards = protocol.InfosToCards(payload.LandlordCards)
		// 更新玩家是否是地主
		for i, p := range m.players {
			m.players[i].IsLandlord = (p.ID == payload.PlayerID)
		}
		if payload.PlayerID == m.playerID {
			m.isLandlord = true
		}

	case protocol.MsgPlayTurn:
		var payload protocol.PlayTurnPayload
		json.Unmarshal(msg.Payload, &payload)
		m.phase = PhasePlaying
		m.currentTurn = payload.PlayerID
		m.mustPlay = payload.MustPlay
		m.canBeat = payload.CanBeat
		m.resetBell() // 重置提示音状态
		if payload.PlayerID == m.playerID {
			if payload.MustPlay {
				m.input.Placeholder = "你必须出牌 (如 33344)"
			} else if payload.CanBeat {
				m.input.Placeholder = "出牌或 PASS"
			} else {
				m.input.Placeholder = "没有能打过的牌，输入 PASS"
			}
			m.input.Focus()
		}
		m.timer = timer.NewWithInterval(time.Duration(payload.Timeout)*time.Second, time.Second)
		return m.timer.Start()

	case protocol.MsgCardPlayed:
		var payload protocol.CardPlayedPayload
		json.Unmarshal(msg.Payload, &payload)
		m.lastPlayedBy = payload.PlayerID
		m.lastPlayedName = payload.PlayerName
		m.lastPlayed = protocol.InfosToCards(payload.Cards)
		m.lastHandType = payload.HandType
		// 更新玩家手牌数
		for i, p := range m.players {
			if p.ID == payload.PlayerID {
				m.players[i].CardsCount = payload.CardsLeft
				break
			}
		}
		// 如果是自己出的牌，从手牌中移除
		if payload.PlayerID == m.playerID {
			m.hand = card.RemoveCards(m.hand, m.lastPlayed)
		}

	case protocol.MsgPlayerPass:
		var payload protocol.PlayerPassPayload
		json.Unmarshal(msg.Payload, &payload)
		// 可以显示 PASS 信息

	case protocol.MsgGameOver:
		var payload protocol.GameOverPayload
		json.Unmarshal(msg.Payload, &payload)
		m.phase = PhaseGameOver
		m.winner = payload.WinnerName
		m.winnerIsLandlord = payload.IsLandlord
		m.input.Placeholder = "按回车返回大厅"

	case protocol.MsgError:
		var payload protocol.ErrorPayload
		json.Unmarshal(msg.Payload, &payload)
		m.error = payload.Message

	case protocol.MsgReconnected:
		var payload protocol.ReconnectedPayload
		json.Unmarshal(msg.Payload, &payload)
		m.playerID = payload.PlayerID
		m.playerName = payload.PlayerName
		if payload.RoomCode != "" {
			m.roomCode = payload.RoomCode
			// 恢复游戏状态
			if payload.GameState != nil {
				m.restoreGameState(payload.GameState)
			} else {
				m.phase = PhaseWaiting
			}
		} else {
			m.phase = PhaseLobby
		}

	case protocol.MsgPlayerOffline:
		var payload protocol.PlayerOfflinePayload
		json.Unmarshal(msg.Payload, &payload)
		// 标记玩家离线
		for i, p := range m.players {
			if p.ID == payload.PlayerID {
				m.players[i].Online = false
				break
			}
		}

	case protocol.MsgPlayerOnline:
		var payload protocol.PlayerOnlinePayload
		json.Unmarshal(msg.Payload, &payload)
		// 标记玩家上线
		for i, p := range m.players {
			if p.ID == payload.PlayerID {
				m.players[i].Online = true
				break
			}
		}

	case protocol.MsgPong:
		var payload protocol.PongPayload
		json.Unmarshal(msg.Payload, &payload)
		m.latency = time.Now().UnixMilli() - payload.ClientTimestamp
	}

	return nil
}

// parseCardsInput 解析出牌输入
func (m *OnlineModel) parseCardsInput(input string) ([]card.Card, error) {
	return card.FindCardsInHand(m.hand, strings.ToUpper(input))
}

// sortHand 排序手牌
func (m *OnlineModel) sortHand() {
	sort.Slice(m.hand, func(i, j int) bool {
		return m.hand[i].Rank > m.hand[j].Rank
	})
}

// resetGameState 重置游戏状态
func (m *OnlineModel) resetGameState() {
	m.roomCode = ""
	m.players = nil
	m.hand = nil
	m.landlordCards = nil
	m.currentTurn = ""
	m.lastPlayedBy = ""
	m.lastPlayed = nil
	m.isLandlord = false
	m.winner = ""
	m.input.Placeholder = "1=创建房间, 2=加入房间, 3=快速匹配"
}

// restoreGameState 从重连数据恢复游戏状态
func (m *OnlineModel) restoreGameState(gs *protocol.GameStateDTO) {
	m.players = gs.Players
	m.hand = protocol.InfosToCards(gs.Hand)
	m.sortHand()
	m.landlordCards = protocol.InfosToCards(gs.LandlordCards)
	m.currentTurn = gs.CurrentTurn
	m.lastPlayed = protocol.InfosToCards(gs.LastPlayed)
	m.lastPlayedBy = gs.LastPlayerID
	m.mustPlay = gs.MustPlay
	m.canBeat = gs.CanBeat

	// 找出自己是否是地主
	for _, p := range m.players {
		if p.ID == m.playerID && p.IsLandlord {
			m.isLandlord = true
			break
		}
	}

	// 根据阶段设置 phase
	switch gs.Phase {
	case "bidding":
		m.phase = PhaseBidding
	case "playing":
		m.phase = PhasePlaying
	case "ended":
		m.phase = PhaseGameOver
	default:
		m.phase = PhaseWaiting
	}
}

// shouldPlayBell 判断是否应该播放提示音
func (m *OnlineModel) shouldPlayBell() bool {
	// 已经播放过了
	if m.bellPlayed {
		return false
	}

	// 必须是自己的回合
	isMyTurn := (m.phase == PhaseBidding && m.bidTurn == m.playerID) ||
		(m.phase == PhasePlaying && m.currentTurn == m.playerID)
	if !isMyTurn {
		return false
	}

	// 检查剩余时间是否为 10 秒
	remaining := m.timer.Timeout
	return remaining <= 10*time.Second && remaining > 9*time.Second
}

// playBell 播放终端提示音
func (m *OnlineModel) playBell() tea.Cmd {
	return tea.Printf("\a") // 发送 ASCII Bell 字符
}

// resetBell 重置提示音状态（新回合开始时调用）
func (m *OnlineModel) resetBell() {
	m.bellPlayed = false
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

// --- 视图渲染 ---

func (m *OnlineModel) connectingView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render("🔌 正在连接服务器...")
}

func (m *OnlineModel) lobbyView() string {
	var sb strings.Builder

	title := titleStyle("🎮 斗地主 - 联网对战")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if m.playerName != "" {
		sb.WriteString(fmt.Sprintf("欢迎, %s!\n\n", m.playerName))
	}

	menu := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		"请选择:",
		"",
		"  1. 创建房间",
		"  2. 加入房间",
		"  3. 快速匹配",
	))
	sb.WriteString(menu)
	sb.WriteString("\n\n")

	m.input.Placeholder = "输入选项 (1/2/3) 或房间号"
	sb.WriteString(m.input.View())

	if m.error != "" {
		sb.WriteString("\n" + errorStyle.Render(m.error))
	}

	return sb.String()
}

func (m *OnlineModel) matchingView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render("🔍 正在匹配中...\n\n按 ESC 取消")
}

func (m *OnlineModel) waitingView() string {
	var sb strings.Builder

	title := titleStyle(fmt.Sprintf("🏠 房间: %s", m.roomCode))
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	// 玩家列表
	var playerList strings.Builder
	playerList.WriteString("玩家列表:\n")
	for _, p := range m.players {
		readyStr := "❌"
		if p.Ready {
			readyStr = "✅"
		}
		meStr := ""
		if p.ID == m.playerID {
			meStr = " (你)"
		}
		playerList.WriteString(fmt.Sprintf("  %s %s%s\n", readyStr, p.Name, meStr))
	}
	playerList.WriteString(fmt.Sprintf("\n等待玩家: %d/3", len(m.players)))

	sb.WriteString(boxStyle.Render(playerList.String()))
	sb.WriteString("\n\n")

	sb.WriteString(m.input.View())

	if m.error != "" {
		sb.WriteString("\n" + errorStyle.Render(m.error))
	}

	return sb.String()
}

func (m *OnlineModel) gameView() string {
	var sb strings.Builder

	// 顶部：底牌和记牌器
	landlordCardsView := m.renderLandlordCardsOnline()
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, landlordCardsView))
	sb.WriteString("\n")

	// 中部：其他玩家信息和上家出牌
	middleSection := m.renderMiddleSection()
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, middleSection))
	sb.WriteString("\n")

	// 底部：自己的手牌和输入
	myHand := m.renderPlayerHand(m.hand)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, myHand))
	sb.WriteString("\n")

	// 提示和输入
	prompt := m.renderPrompt()
	sb.WriteString(prompt)

	if m.error != "" {
		sb.WriteString("\n" + errorStyle.Render(m.error))
	}

	return sb.String()
}

func (m *OnlineModel) gameOverView() string {
	winnerType := "农民"
	if m.winnerIsLandlord {
		winnerType = "地主"
	}

	msg := fmt.Sprintf("🎮 游戏结束!\n\n🏆 %s (%s) 获胜!\n\n按回车返回大厅", m.winner, winnerType)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(msg)
}

func (m *OnlineModel) renderLandlordCardsOnline() string {
	if len(m.landlordCards) == 0 {
		return boxStyle.Render("底牌: (待揭晓)")
	}

	// 渲染底牌
	var cardStrs []string
	for _, c := range m.landlordCards {
		style := blackStyle
		if c.Color == card.Red {
			style = redStyle
		}
		cardStrs = append(cardStrs, style.Render(fmt.Sprintf("%s%s", c.Suit.String(), c.Rank.String())))
	}

	content := "底牌: " + strings.Join(cardStrs, " ")
	return boxStyle.Render(content)
}

func (m *OnlineModel) renderMiddleSection() string {
	// 渲染其他玩家和上家出牌
	var parts []string

	// 其他玩家信息
	for _, p := range m.players {
		if p.ID == m.playerID {
			continue
		}

		icon := FarmerIcon
		if p.IsLandlord {
			icon = LandlordIcon
		}

		nameStyle := lipgloss.NewStyle()
		if m.currentTurn == p.ID {
			nameStyle = nameStyle.Foreground(lipgloss.Color("220")).Bold(true)
		}

		info := fmt.Sprintf("%s %s\n🃏 %d张", icon, nameStyle.Render(p.Name), p.CardsCount)
		parts = append(parts, boxStyle.Width(15).Render(info))
	}

	// 上家出牌
	lastPlayView := "(等待出牌...)"
	if len(m.lastPlayed) > 0 {
		var cardStrs []string
		for _, c := range m.lastPlayed {
			style := blackStyle
			if c.Color == card.Red {
				style = redStyle
			}
			cardStrs = append(cardStrs, style.Render(c.Rank.String()))
		}
		lastPlayView = fmt.Sprintf("%s: %s\n%s", m.lastPlayedName, strings.Join(cardStrs, " "), m.lastHandType)
	}
	parts = append(parts, boxStyle.Width(25).Render(lastPlayView))

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m *OnlineModel) renderPrompt() string {
	var sb strings.Builder

	if m.phase == PhaseBidding {
		if m.bidTurn == m.playerID {
			sb.WriteString(fmt.Sprintf("⏳ %s | 轮到你叫地主!\n", m.timer.View()))
		} else {
			for _, p := range m.players {
				if p.ID == m.bidTurn {
					sb.WriteString(fmt.Sprintf("等待 %s 叫地主...\n", p.Name))
					break
				}
			}
		}
	} else if m.phase == PhasePlaying {
		if m.currentTurn == m.playerID {
			icon := FarmerIcon
			if m.isLandlord {
				icon = LandlordIcon
			}
			sb.WriteString(fmt.Sprintf("⏳ %s | 轮到你出牌! %s\n", m.timer.View(), icon))
		} else {
			for _, p := range m.players {
				if p.ID == m.currentTurn {
					sb.WriteString(fmt.Sprintf("等待 %s 出牌...\n", p.Name))
					break
				}
			}
		}
	}

	sb.WriteString(m.input.View())

	return promptStyle.Render(sb.String())
}

// renderPlayerHand 渲染玩家手牌（复用原有代码）
func (m *OnlineModel) renderPlayerHand(hand []card.Card) string {
	if len(hand) == 0 {
		return boxStyle.Render("(无手牌)")
	}

	// 简化版手牌显示
	var rankStr, suitStr strings.Builder
	for _, c := range hand {
		style := blackStyle
		if c.Color == card.Red {
			style = redStyle
		}
		style = style.Align(lipgloss.Center).Margin(0, 1)
		rankStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Rank.String())))
		suitStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Suit.String())))
	}

	icon := FarmerIcon
	if m.isLandlord {
		icon = LandlordIcon
	}
	title := fmt.Sprintf("我的手牌 %s (%d张)", icon, len(hand))
	content := lipgloss.JoinVertical(lipgloss.Center, title, rankStr.String(), suitStr.String())
	return boxStyle.Render(content)
}

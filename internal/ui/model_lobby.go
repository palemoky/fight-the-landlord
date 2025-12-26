package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/palemoky/fight-the-landlord/internal/network/client"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// LobbyModel handles the lobby interface (Menu, Room List, Leaderboard, Stats)
type LobbyModel struct {
	client *client.Client
	width  int
	height int

	// Navigation
	selectedIndex int // Menu index

	// Data
	onlineCount     int
	availableRooms  []protocol.RoomListItem
	selectedRoomIdx int
	leaderboard     []protocol.LeaderboardEntry
	myStats         *protocol.StatsResultPayload

	// Chat
	chatHistory []string
	chatInput   textinput.Model

	// Input
	input *textinput.Model
}

func NewLobbyModel(c *client.Client, input *textinput.Model) *LobbyModel {
	chatInput := textinput.New()
	chatInput.Placeholder = "按 / 键聊天..."
	chatInput.CharLimit = 50
	chatInput.Width = 30

	return &LobbyModel{
		client:    c,
		input:     input,
		chatInput: chatInput,
	}
}

func (m *LobbyModel) Init() tea.Cmd {
	return nil
}

func (m *LobbyModel) View() string {
	return "" // Not used directly, managed by OnlineModel
}

// Update handles lobby-specific updates
func (m *LobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// The parent OnlineModel handles global keys (Esc, etc.) and server messages
	// This Update method is mainly for internal component updates if needed
	return m, nil
}

// View Logic moved from online_views.go

func (m *LobbyModel) lobbyView(onlineModel *OnlineModel) string {
	var sb strings.Builder

	title := titleStyle("🎮 欢乐斗地主")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if onlineModel.playerName != "" {
		welcome := fmt.Sprintf("欢迎, %s!", onlineModel.playerName)
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, welcome))
		sb.WriteString("\n")

		// Display online count
		if m.onlineCount > 0 {
			onlineInfo := fmt.Sprintf("🌐 在线玩家: %d 人", m.onlineCount)
			onlineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
			sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, onlineStyle.Render(onlineInfo)))
		}
		sb.WriteString("\n")

		// Reconnect status handled by OnlineModel, passed in or handled by parent view composition
		if onlineModel.reconnecting || onlineModel.reconnectSuccess {
			var reconnectStyle lipgloss.Style
			if onlineModel.reconnectSuccess {
				reconnectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
			} else {
				reconnectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			}
			sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, reconnectStyle.Render(onlineModel.reconnectMessage)))
		}
		sb.WriteString("\n")
	}

	menuItems := []string{
		"1. 快速匹配",
		"2. 创建房间",
		"3. 加入房间",
		"4. 排行榜",
		"5. 我的战绩",
		"6. 游戏规则",
	}

	var menuLines []string
	menuLines = append(menuLines, "请选择:", "")
	for i, item := range menuItems {
		prefix := "  "
		if i == m.selectedIndex {
			prefix = "▶ "
		}
		menuLines = append(menuLines, prefix+item)
	}

	// Build menu box
	menu := boxStyle.Padding(0, 2).Render(lipgloss.JoinVertical(lipgloss.Left, menuLines...))
	menuHeight := lipgloss.Height(menu)

	// Build chat box with input at bottom
	var chatLines []string
	if len(m.chatHistory) > 0 {
		count := len(m.chatHistory)
		start := 0
		if count > 5 {
			start = count - 5
		}
		for i := start; i < count; i++ {
			chatLines = append(chatLines, m.chatHistory[i])
		}
	} else {
		chatLines = append(chatLines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("暂无消息..."))
	}

	// Chat input view
	chatInputView := m.chatInput.View()
	if !m.chatInput.Focused() {
		chatInputView = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("按 / 键聊天...")
	}

	// Chat header
	chatHeader := lipgloss.NewStyle().Bold(true).Render("💬 聊天室")

	// Calculate inner height (menu height - 2 for top/bottom border)
	innerHeight := menuHeight - 2

	// Build content: header + messages
	contentLines := []string{chatHeader}
	contentLines = append(contentLines, chatLines...)

	// Calculate how many empty lines needed to push input to bottom
	// innerHeight = header(1) + messages + empty_lines + input(1)
	usedLines := len(contentLines) + 1 // +1 for input line
	emptyLines := innerHeight - usedLines
	if emptyLines < 0 {
		emptyLines = 0
	}

	// Add empty lines as spacer
	for i := 0; i < emptyLines; i++ {
		contentLines = append(contentLines, "")
	}
	// Add input at bottom
	contentLines = append(contentLines, chatInputView)

	chatBoxContent := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	chatBox := boxStyle.Width(30).Height(innerHeight).Render(chatBoxContent)

	// Place menu and chat side by side
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, menu, "  ", chatBox)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, mainContent))
	sb.WriteString("\n\n")

	m.input.Placeholder = "↑↓ 选择 | 回车确认 | 或输入房间号"
	inputView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, m.input.View())
	sb.WriteString(inputView)

	if onlineModel.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(onlineModel.error))
		sb.WriteString(errorView)
	}

	return sb.String()
}

func (m *LobbyModel) roomListView(onlineModel *OnlineModel) string {
	var sb strings.Builder

	title := titleStyle("📋 可加入的房间")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if len(m.availableRooms) == 0 {
		noRooms := "暂无可加入的房间\n\n按 ESC 返回大厅"
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, noRooms))
	} else {
		var roomList strings.Builder
		roomList.WriteString("房间列表:\n\n")

		for i, room := range m.availableRooms {
			prefix := "  "
			if i == m.selectedRoomIdx {
				prefix = "▶ "
			}
			roomList.WriteString(fmt.Sprintf("%s房间 %s  (%d/3)\n", prefix, room.RoomCode, room.PlayerCount))
		}

		roomList.WriteString("\n↑↓ 选择  回车加入  ESC 返回")

		roomBox := boxStyle.Render(roomList.String())
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, roomBox))
		sb.WriteString("\n\n")
	}

	inputView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, m.input.View())
	sb.WriteString(inputView)

	if onlineModel.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(onlineModel.error))
		sb.WriteString(errorView)
	}

	return sb.String()
}

func (m *LobbyModel) leaderboardView() string {
	var sb strings.Builder

	title := titleStyle("🏆 排行榜")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if len(m.leaderboard) > 0 {
		// renderLeaderboard internal helper moved here
		leaderboard := m.renderLeaderboardTable()
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, leaderboard))
	} else {
		noData := "正在加载排行榜..."
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, noData))
	}

	sb.WriteString("\n\n")
	hint := "按 ESC 返回大厅"
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hint))

	return sb.String()
}

func (m *LobbyModel) renderLeaderboardTable() string {
	var sb strings.Builder

	// Center the title
	title := "🏆 排行榜 TOP 10"
	titleLine := lipgloss.PlaceHorizontal(50, lipgloss.Center, title)
	sb.WriteString(titleLine + "\n")
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	// Header - use tabs for alignment
	sb.WriteString("排名\t玩家\t\t积分\t胜场\t胜率\n")
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	for _, e := range m.leaderboard {
		rankStr := fmt.Sprintf("%2d.", e.Rank)

		// Use tabs for alignment - works better with Chinese characters
		fmt.Fprintf(&sb, "%s\t%s\t\t%d\t%d\t%.1f%%\n",
			rankStr, truncateName(e.PlayerName, 10), e.Score, e.Wins, e.WinRate)
	}

	return boxStyle.Render(sb.String())
}

func (m *LobbyModel) statsView() string {
	var sb strings.Builder

	title := titleStyle("📊 我的战绩")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if m.myStats != nil && m.myStats.TotalGames > 0 {
		stats := m.renderMyStatsTable()
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, stats))
	} else {
		noData := "暂无战绩数据"
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, noData))
	}

	sb.WriteString("\n\n")
	hint := "按 ESC 返回大厅"
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hint))

	return sb.String()
}

func (m *LobbyModel) renderMyStatsTable() string {
	s := m.myStats
	var sb strings.Builder
	sb.WriteString("📊 我的战绩\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// 排名和积分
	rankStr := "未上榜"
	if s.Rank > 0 {
		rankStr = fmt.Sprintf("#%d", s.Rank)
	}
	fmt.Fprintf(&sb, "排名: %s  |  积分: %d\n", rankStr, s.Score)
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// 总战绩
	fmt.Fprintf(&sb, "总场次: %d  胜: %d  负: %d  胜率: %.1f%%\n",
		s.TotalGames, s.Wins, s.Losses, s.WinRate)

	// 地主/农民分开
	landlordRate := 0.0
	if s.LandlordGames > 0 {
		landlordRate = float64(s.LandlordWins) / float64(s.LandlordGames) * 100
	}
	farmerRate := 0.0
	if s.FarmerGames > 0 {
		farmerRate = float64(s.FarmerWins) / float64(s.FarmerGames) * 100
	}

	fmt.Fprintf(&sb, "地主: %d胜/%d场 (%.1f%%)  |  农民: %d胜/%d场 (%.1f%%)\n",
		s.LandlordWins, s.LandlordGames, landlordRate,
		s.FarmerWins, s.FarmerGames, farmerRate)

	streakStr := ""
	if s.CurrentStreak > 0 {
		streakStr = fmt.Sprintf("🔥 %d 连胜!", s.CurrentStreak)
	} else if s.CurrentStreak < 0 {
		streakStr = fmt.Sprintf("💔 %d 连败", -s.CurrentStreak)
	}
	if s.MaxWinStreak > 0 {
		streakStr += fmt.Sprintf("  最高连胜: %d", s.MaxWinStreak)
	}
	if streakStr != "" {
		sb.WriteString(streakStr + "\n")
	}

	return boxStyle.Render(sb.String())
}

func (m *LobbyModel) handleUpKey(phase GamePhase) {
	if phase == PhaseRoomList && len(m.availableRooms) > 0 {
		m.selectedRoomIdx--
		if m.selectedRoomIdx < 0 {
			m.selectedRoomIdx = len(m.availableRooms) - 1
		}
	} else if phase == PhaseLobby {
		m.selectedIndex--
		if m.selectedIndex < 0 {
			m.selectedIndex = 5
		}
	}
}

func (m *LobbyModel) handleDownKey(phase GamePhase) {
	if phase == PhaseRoomList && len(m.availableRooms) > 0 {
		m.selectedRoomIdx++
		if m.selectedRoomIdx >= len(m.availableRooms) {
			m.selectedRoomIdx = 0
		}
	} else if phase == PhaseLobby {
		m.selectedIndex++
		if m.selectedIndex > 5 {
			m.selectedIndex = 0
		}
	}
}

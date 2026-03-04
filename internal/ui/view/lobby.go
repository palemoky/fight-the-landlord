// Package view provides UI rendering functions.
package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/ui/common"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

// getNotificationStyle returns the appropriate style for a notification type.
func getNotificationStyle(notifyType model.NotificationType) lipgloss.Style {
	switch notifyType {
	case model.NotifyError, model.NotifyRateLimit:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	case model.NotifyReconnecting:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case model.NotifyReconnectSuccess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	case model.NotifyMaintenance:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	case model.NotifyOnlineCount:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	default:
		return lipgloss.NewStyle()
	}
}

// renderChatBox renders the chat box component for the lobby.
func renderChatBox(lobby model.LobbyAccessor, height int) string {
	var chatLines []string
	if len(lobby.ChatHistory()) > 0 {
		history := lobby.ChatHistory()
		count := len(history)
		start := 0
		if count > 5 {
			start = count - 5
		}
		for i := start; i < count; i++ {
			chatLines = append(chatLines, history[i])
		}
	} else {
		chatLines = append(chatLines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("暂无消息..."))
	}

	chatInputView := lobby.ChatInput().View()
	if !lobby.ChatInput().Focused() {
		chatInputView = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("按 / 键聊天...")
	}

	chatHeader := lipgloss.NewStyle().Bold(true).Render("💬 聊天室")
	innerHeight := height - 2
	usedLines := 1 + len(chatLines) + 1 // header + chat + input
	emptyLines := max(innerHeight-usedLines, 0)
	contentLines := make([]string, 0, 1+len(chatLines)+emptyLines+1)
	contentLines = append(contentLines, chatHeader)
	contentLines = append(contentLines, chatLines...)
	for range emptyLines {
		contentLines = append(contentLines, "")
	}
	contentLines = append(contentLines, chatInputView)

	chatBoxWidth := 50
	chatBoxContent := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	return common.BoxStyle.Width(chatBoxWidth).Height(innerHeight).Render(chatBoxContent)
}

// LobbyView renders the lobby view.
func LobbyView(m model.Model) string {
	lobby := m.Lobby()
	var sb strings.Builder

	title := common.TitleStyle("🎮 欢乐斗地主")
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, title))
	sb.WriteString("\n\n")

	if m.PlayerName() != "" {
		welcome := fmt.Sprintf("欢迎, %s!", m.PlayerName())
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, welcome))
		sb.WriteString("\n")

		// Show system notification
		if notification := m.GetCurrentNotification(); notification != nil {
			notificationStyle := getNotificationStyle(notification.Type)
			sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center,
				notificationStyle.Render(notification.Message)))
		}
		sb.WriteString("\n")
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

	lobbyModel := m.Lobby()
	menuLines := make([]string, 0, 2+len(menuItems))
	menuLines = append(menuLines, "请选择:", "")
	for i, item := range menuItems {
		prefix := "  "
		if i == lobbyModel.SelectedIndex() {
			prefix = "▶ "
		}
		menuLines = append(menuLines, prefix+item)
	}

	menu := common.BoxStyle.Padding(0, 2).Render(lipgloss.JoinVertical(lipgloss.Left, menuLines...))
	menuHeight := lipgloss.Height(menu)

	// Chat box
	chatBox := renderChatBox(lobby, menuHeight)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, menu, "  ", chatBox)
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, mainContent))
	sb.WriteString("\n\n")

	// Only show blinking cursor on lobby input when chat is not focused
	var inputView string
	if lobby.ChatInput().Focused() {
		m.Input().Blur()
		inputView = lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("> ↑↓ 选择 | 回车确认 | 或输入房间号"))
	} else {
		m.Input().Focus()
		m.Input().Placeholder = "↑↓ 选择 | 回车确认 | 或输入房间号"
		inputView = lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, m.Input().View())
	}
	sb.WriteString(inputView)

	creditStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	credit := creditStyle.Render("Made with ❤️ by Palemoky")
	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, credit))

	content := sb.String()
	return lipgloss.Place(m.Width(), m.Height(), lipgloss.Center, lipgloss.Center, content)
}

// RoomListView renders the room list view.
func RoomListView(m model.Model) string {
	lobby := m.Lobby()
	var sb strings.Builder

	title := common.TitleStyle("📋 可加入的房间")
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, title))
	sb.WriteString("\n\n")

	rooms := lobby.AvailableRooms()
	if len(rooms) == 0 {
		noRooms := "暂无可加入的房间\n\n按 ESC 返回大厅"
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, noRooms))
	} else {
		var roomList strings.Builder
		roomList.WriteString("房间列表:\n\n")

		for i, room := range rooms {
			prefix := "  "
			if i == lobby.SelectedRoomIdx() {
				prefix = "▶ "
			}
			fmt.Fprintf(&roomList, "%s房间 %s  (%d/3)\n", prefix, room.RoomCode, room.PlayerCount)
		}

		roomList.WriteString("\n↑↓ 选择  回车加入  ESC 返回")

		roomBox := common.BoxStyle.Render(roomList.String())
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, roomBox))
		sb.WriteString("\n\n")
	}

	inputView := lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, m.Input().View())
	sb.WriteString(inputView)

	return sb.String()
}

// LeaderboardView renders the leaderboard view.
func LeaderboardView(m model.Model) string {
	lobby := m.Lobby()
	var sb strings.Builder

	title := common.TitleStyle("🏆 排行榜")
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, title))
	sb.WriteString("\n\n")

	entries := lobby.Leaderboard()
	if len(entries) > 0 {
		leaderboard := renderLeaderboardTable(entries)
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, leaderboard))
	} else {
		noData := "正在加载排行榜..."
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, noData))
	}

	sb.WriteString("\n\n")
	hint := "按 ESC 返回大厅"
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, hint))

	return sb.String()
}

func renderLeaderboardTable(entries []protocol.LeaderboardEntry) string {
	var sb strings.Builder

	title := "🏆 排行榜 TOP 10"
	titleLine := lipgloss.PlaceHorizontal(50, lipgloss.Center, title)
	sb.WriteString(titleLine + "\n")
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	sb.WriteString("排名\t玩家\t\t积分\t胜场\t胜率\n")
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	for _, e := range entries {
		rankStr := fmt.Sprintf("%2d.", e.Rank)
		fmt.Fprintf(&sb, "%s\t%s\t\t%d\t%d\t%.1f%%\n",
			rankStr, common.TruncateName(e.PlayerName, 10), e.Score, e.Wins, e.WinRate)
	}

	return common.BoxStyle.Render(sb.String())
}

// StatsView renders the stats view.
func StatsView(m model.Model) string {
	lobby := m.Lobby()
	var sb strings.Builder

	title := common.TitleStyle("📊 我的战绩")
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, title))
	sb.WriteString("\n\n")

	stats := lobby.MyStats()
	if stats != nil && stats.TotalGames > 0 {
		statsTable := renderStatsTable(stats)
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, statsTable))
	} else {
		noData := "暂无战绩数据"
		sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, noData))
	}

	sb.WriteString("\n\n")
	hint := "按 ESC 返回大厅"
	sb.WriteString(lipgloss.PlaceHorizontal(m.Width(), lipgloss.Center, hint))

	return sb.String()
}

func renderStatsTable(s *protocol.StatsResultPayload) string {
	var sb strings.Builder
	sb.WriteString("📊 我的战绩\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	rankStr := "未上榜"
	if s.Rank > 0 {
		rankStr = fmt.Sprintf("#%d", s.Rank)
	}
	fmt.Fprintf(&sb, "排名: %s  |  积分: %d\n", rankStr, s.Score)
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	fmt.Fprintf(&sb, "总场次: %d  胜: %d  负: %d  胜率: %.1f%%\n",
		s.TotalGames, s.Wins, s.Losses, s.WinRate)

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

	return common.BoxStyle.Render(sb.String())
}

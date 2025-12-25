package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/palemoky/fight-the-landlord/internal/card"
)

// --- 视图渲染 ---

func (m *OnlineModel) connectingView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render("🔌 正在连接服务器...")
}

func (m *OnlineModel) lobbyView() string {
	var sb strings.Builder

	title := titleStyle("🎮 欢乐斗地主")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if m.playerName != "" {
		welcome := fmt.Sprintf("欢迎, %s!", m.playerName)
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, welcome))
		sb.WriteString("\n\n")
	}

	menu := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		"请选择:",
		"",
		"  1. 创建房间",
		"  2. 加入房间",
		"  3. 快速匹配",
		"  4. 排行榜",
		"  5. 我的战绩",
	))
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, menu))
	sb.WriteString("\n\n")

	m.input.Placeholder = "输入选项 (1-5) 或房间号"
	inputView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, m.input.View())
	sb.WriteString(inputView)

	if m.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(m.error))
		sb.WriteString(errorView)
	}

	return sb.String()
}

// renderLeaderboard 渲染排行榜
func (m *OnlineModel) renderLeaderboard() string {
	var sb strings.Builder
	sb.WriteString("🏆 排行榜 TOP 10\n")
	sb.WriteString(strings.Repeat("─", 50) + "\n")
	sb.WriteString(fmt.Sprintf("%-4s %-12s %8s %6s %8s\n", "排名", "玩家", "积分", "胜场", "胜率"))
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	for _, e := range m.leaderboard {
		rankIcon := ""
		switch e.Rank {
		case 1:
			rankIcon = "🥇"
		case 2:
			rankIcon = "🥈"
		case 3:
			rankIcon = "🥉"
		default:
			rankIcon = fmt.Sprintf("%2d.", e.Rank)
		}
		sb.WriteString(fmt.Sprintf("%-4s %-12s %8d %6d %7.1f%%\n",
			rankIcon, truncateName(e.PlayerName, 10), e.Score, e.Wins, e.WinRate))
	}

	return boxStyle.Render(sb.String())
}

// renderMyStats 渲染我的战绩
func (m *OnlineModel) renderMyStats() string {
	s := m.myStats
	var sb strings.Builder
	sb.WriteString("📊 我的战绩\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// 排名和积分
	rankStr := "未上榜"
	if s.Rank > 0 {
		rankStr = fmt.Sprintf("#%d", s.Rank)
	}
	sb.WriteString(fmt.Sprintf("排名: %s  |  积分: %d\n", rankStr, s.Score))
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// 总战绩
	sb.WriteString(fmt.Sprintf("总场次: %d  胜: %d  负: %d  胜率: %.1f%%\n",
		s.TotalGames, s.Wins, s.Losses, s.WinRate))

	// 地主/农民分开
	landlordRate := 0.0
	if s.LandlordGames > 0 {
		landlordRate = float64(s.LandlordWins) / float64(s.LandlordGames) * 100
	}
	farmerRate := 0.0
	if s.FarmerGames > 0 {
		farmerRate = float64(s.FarmerWins) / float64(s.FarmerGames) * 100
	}

	sb.WriteString(fmt.Sprintf("地主: %d胜/%d场 (%.1f%%)  |  农民: %d胜/%d场 (%.1f%%)\n",
		s.LandlordWins, s.LandlordGames, landlordRate,
		s.FarmerWins, s.FarmerGames, farmerRate))

	// 连胜信息
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

func (m *OnlineModel) leaderboardView() string {
	var sb strings.Builder

	title := titleStyle("🏆 排行榜")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if len(m.leaderboard) > 0 {
		leaderboard := m.renderLeaderboard()
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

func (m *OnlineModel) statsView() string {
	var sb strings.Builder

	title := titleStyle("📊 我的战绩")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if m.myStats != nil && m.myStats.TotalGames > 0 {
		stats := m.renderMyStats()
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

func (m *OnlineModel) matchingView() string {
	elapsed := ""
	if !m.matchingStartTime.IsZero() {
		seconds := int(time.Since(m.matchingStartTime).Seconds())
		elapsed = fmt.Sprintf("\n已等待: %d 秒", seconds)
	}

	content := fmt.Sprintf("🔍 正在匹配中...%s\n\n按 ESC 取消", elapsed)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(content)
}

func (m *OnlineModel) roomListView() string {
	var sb strings.Builder

	title := titleStyle("📋 可加入的房间")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	if len(m.availableRooms) == 0 {
		noRooms := "暂无可加入的房间\n\n按 ESC 返回大厅"
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, noRooms))
	} else {
		// 显示房间列表
		var roomList strings.Builder
		roomList.WriteString("房间列表:\n\n")

		for i, room := range m.availableRooms {
			prefix := "  "
			if i == m.selectedRoomIndex {
				prefix = "▶ " // 选中标记
			}
			roomList.WriteString(fmt.Sprintf("%s房间 %s  (%d/3)\n", prefix, room.RoomCode, room.PlayerCount))
		}

		roomList.WriteString("\n↑↓ 选择  回车加入  ESC 返回")

		roomBox := boxStyle.Render(roomList.String())
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, roomBox))
		sb.WriteString("\n\n")
	}

	// 输入框用于直接输入房间号
	inputView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, m.input.View())
	sb.WriteString(inputView)

	if m.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(m.error))
		sb.WriteString(errorView)
	}

	return sb.String()
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

	playerBox := boxStyle.Render(playerList.String())
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, playerBox))
	sb.WriteString("\n\n")

	inputView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, m.input.View())
	sb.WriteString(inputView)

	if m.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(m.error))
		sb.WriteString(errorView)
	}

	return sb.String()
}

func (m *OnlineModel) gameView() string {
	var sb strings.Builder

	// 顶部：底牌和记牌器
	landlordCardsView := m.renderLandlordCardsOnline()
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, landlordCardsView))
	sb.WriteString("\n")

	// 记牌器（如果启用）
	if m.cardCounterEnabled {
		cardCounter := m.renderCardCounter()
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, cardCounter))
		sb.WriteString("\n")
	}

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

func (m *OnlineModel) renderCardCounter() string {
	if !m.cardCounterEnabled {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("记牌器 (C)\n")
	sb.WriteString(strings.Repeat("─", 44) + "\n")

	// 定义牌的顺序：大王 小王 2 A K Q J 10 9 8 7 6 5 4 3
	ranks := []card.Rank{
		card.RankRedJoker, card.RankBlackJoker, card.Rank2,
		card.RankA, card.RankK, card.RankQ, card.RankJ, card.Rank10,
		card.Rank9, card.Rank8, card.Rank7, card.Rank6,
		card.Rank5, card.Rank4, card.Rank3,
	}

	// 第一行：牌名
	var names []string
	for _, rank := range ranks {
		name := rank.String()
		switch rank {
		case card.RankRedJoker:
			name = "R"
		case card.RankBlackJoker:
			name = "B"
		}
		names = append(names, fmt.Sprintf("%-2s", name))
	}
	sb.WriteString(strings.Join(names, "│") + "\n")
	sb.WriteString(strings.Repeat("─", 44) + "\n")

	// 第二行：剩余数量
	var counts []string
	for _, rank := range ranks {
		count := m.remainingCards[rank]
		counts = append(counts, fmt.Sprintf("%-2d", count))
	}
	sb.WriteString(strings.Join(counts, "│"))

	return boxStyle.Render(sb.String())
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
		// 反向遍历，从大到小显示
		for i := len(m.lastPlayed) - 1; i >= 0; i-- {
			c := m.lastPlayed[i]
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

// renderPlayerHand 渲染玩家手牌
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

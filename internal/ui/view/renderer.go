// Package view provides UI rendering functions.
package view

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	gameClient "github.com/palemoky/fight-the-landlord/internal/client"
	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/ui/common"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

// CreateViewRenderer creates a view renderer function that can be injected into OnlineModel.
func CreateViewRenderer() func(model.Model, model.GamePhase) string {
	return func(m model.Model, phase model.GamePhase) string {
		switch phase {
		case model.PhaseLobby:
			return LobbyView(m)
		case model.PhaseRoomList:
			return RoomListView(m)
		case model.PhaseWaiting:
			return WaitingView(m)
		case model.PhaseBidding, model.PhasePlaying:
			return GameView(m)
		case model.PhaseGameOver:
			return GameOverView(m)
		case model.PhaseLeaderboard:
			return LeaderboardView(m)
		case model.PhaseStats:
			return StatsView(m)
		case model.PhaseRules:
			return RulesView(m.Width(), m.Height())
		default:
			return "Unknown phase"
		}
	}
}

// WaitingView renders the waiting room view.
func WaitingView(m model.Model) string {
	width := m.Width()
	game := m.Game()
	state := game.State()

	var sb strings.Builder

	title := common.TitleStyle(fmt.Sprintf("🏠 房间: %s", state.RoomCode))
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	var playerList strings.Builder
	playerList.WriteString("玩家列表:\n")
	for _, p := range state.Players {
		readyStr := "❌"
		if p.Ready {
			readyStr = "✅"
		}
		meStr := ""
		if p.ID == m.PlayerID() {
			meStr = " (你)"
		} else if p.IsBot {
			meStr = " (AI)"
		}
		fmt.Fprintf(&playerList, "  %s %s%s\n", readyStr, p.Name, meStr)
	}
	fmt.Fprintf(&playerList, "\n等待玩家: %d/3", len(state.Players))

	playerBox := common.BoxStyle.Render(playerList.String())
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, playerBox))
	sb.WriteString("\n\n")

	inputView := lipgloss.PlaceHorizontal(width, lipgloss.Center, m.Input().View())
	sb.WriteString(inputView)

	// Chat Rendering
	chatBox := RenderChatBox(game.ChatHistory())
	if chatBox != "" {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, chatBox))
	}

	return sb.String()
}

// GameView renders the main game view for bidding and playing phases.
func GameView(m model.Model) string {
	width := m.Width()
	height := m.Height()
	game := m.Game()
	state := game.State()
	playerID := m.PlayerID()

	var sb strings.Builder

	// Top section - landlord cards and card counter
	topSection := renderTopSection(state, game.CardCounterEnabled())
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, topSection))
	sb.WriteString("\n")

	// Middle section - other players info and last played cards
	middleSection := renderMiddleSection(state, playerID)
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, middleSection))
	sb.WriteString("\n")

	// Player hand
	myHand := renderPlayerHand(state.Hand, state.IsLandlord)
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, myHand))
	sb.WriteString("\n")

	// Prompt with timer
	prompt := renderPrompt(m, game, state, playerID)
	sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, prompt))

	// Chat Rendering
	chatBox := RenderChatBox(game.ChatHistory())
	if chatBox != "" {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, chatBox))
	}

	gameContent := sb.String()

	// Overlays
	if game.ShowQuickMsgMenu() {
		menuContent := renderQuickMsgMenu(game.QuickMsgScroll(), game.QuickMsgInput())
		return lipgloss.Place(width, height,
			lipgloss.Center, lipgloss.Center,
			menuContent,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	if game.ShowingHelp() {
		helpContent := RenderGameRules()
		return lipgloss.Place(width, height,
			lipgloss.Center, lipgloss.Center,
			helpContent,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, gameContent)
}

// GameOverView renders the game over view.
func GameOverView(m model.Model) string {
	width := m.Width()
	state := m.Game().State()

	winnerType := "农民"
	if state.WinnerIsLandlord {
		winnerType = "地主"
	}

	winnerName := state.Winner
	// Find winner name from players
	for _, p := range state.Players {
		if p.ID == state.Winner {
			winnerName = p.Name
			break
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "🎮 游戏结束!\n\n🏆 %s (%s) 获胜!\n", winnerName, winnerType)
	if state.FinalMultiplier > 0 {
		fmt.Fprintf(&sb, "\n💥 本局倍数: ×%d\n", state.FinalMultiplier)
	}
	if len(state.Scores) > 0 {
		sb.WriteString("\n── 本局得分 ──\n")
		for _, s := range state.Scores {
			role := "农民"
			if s.IsLandlord {
				role = "地主"
			}
			fmt.Fprintf(&sb, "%s (%s): %+d\n", s.PlayerName, role, s.Score)
		}
	}
	sb.WriteString("\n按 ESC 返回大厅")

	content := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(sb.String())

	return lipgloss.Place(width, m.Height(), lipgloss.Center, lipgloss.Center, content)
}

// --- Helper rendering functions ---

func renderTopSection(state *gameClient.GameState, cardCounterEnabled bool) string {
	bottomCardsView := renderBottomCards(state.BottomCards)
	if cardCounterEnabled && state.CardCounter != nil {
		cardCounter := renderCardCounter(state.CardCounter)
		return lipgloss.JoinHorizontal(lipgloss.Top, cardCounter, "  ", bottomCardsView)
	}
	return bottomCardsView
}

func renderBottomCards(bottomCards []card.Card) string {
	if len(bottomCards) == 0 {
		return common.BoxStyle.Render("底牌: (待揭晓)")
	}

	var rankStr, suitStr strings.Builder
	for _, c := range bottomCards {
		style := common.BlackStyle
		if c.Color == card.Red {
			style = common.RedStyle
		}
		style = style.Align(lipgloss.Center).Margin(0, 1)
		rankStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Rank.String())))
		suitStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Suit.String())))
	}

	title := "底牌"
	content := lipgloss.JoinVertical(lipgloss.Center, title, rankStr.String(), suitStr.String())
	return common.BoxStyle.Render(content)
}

func renderCardCounter(counter *gameClient.CardCounter) string {
	if counter == nil {
		return ""
	}

	var sb strings.Builder
	ranks := []card.Rank{
		card.RankRedJoker, card.RankBlackJoker, card.Rank2,
		card.RankA, card.RankK, card.RankQ, card.RankJ, card.Rank10,
		card.Rank9, card.Rank8, card.Rank7, card.Rank6,
		card.Rank5, card.Rank4, card.Rank3,
	}

	names := make([]string, 0, len(ranks))
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

	remaining := counter.GetRemaining()
	counts := make([]string, 0, len(ranks))
	for _, rank := range ranks {
		count := remaining[rank]
		counts = append(counts, fmt.Sprintf("%-2d", count))
	}
	sb.WriteString(strings.Join(counts, "│"))

	return common.BoxStyle.Render(sb.String())
}

func renderMiddleSection(state *gameClient.GameState, myPlayerID string) string {
	parts := make([]string, 0, 3) // max 2 other players + 1 last play view
	for _, p := range state.Players {
		if p.ID == myPlayerID {
			continue
		}

		icon := common.FarmerIcon
		if p.IsLandlord {
			icon = common.LandlordIcon
		}
		if p.IsBot {
			icon = "🤖"
		}

		nameStyle := lipgloss.NewStyle()
		if state.CurrentTurn == p.ID {
			nameStyle = nameStyle.Foreground(lipgloss.Color("220")).Bold(true)
		}

		info := fmt.Sprintf("%s %s\n🃏 %d张", icon, nameStyle.Render(p.Name), p.CardsCount)
		parts = append(parts, common.BoxStyle.Width(15).Render(info))
	}

	lastPlayView := "(等待出牌...)"
	if len(state.LastPlayed) > 0 {
		lastPlayView = renderLastPlayed(state)
	}
	// 宽度随出牌长度自适应：以原宽度为下限，20 张牌（点数最长 "10"）的极限宽度为上限
	boxWidth := min(max(25, lipgloss.Width(lastPlayView)), 62)
	parts = append(parts, common.BoxStyle.Width(boxWidth).Render(lastPlayView))

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func renderLastPlayed(state *gameClient.GameState) string {
	var cardStrs []string
	for i := len(state.LastPlayed) - 1; i >= 0; i-- {
		c := state.LastPlayed[i]
		style := common.BlackStyle
		if c.Color == card.Red {
			style = common.RedStyle
		}
		cardStrs = append(cardStrs, style.Render(c.Rank.String()))
	}

	// 上家角色图标
	icon := common.FarmerIcon
	for _, p := range state.Players {
		if p.ID == state.LastPlayedBy {
			if p.IsLandlord {
				icon = common.LandlordIcon
			}
			if p.IsBot {
				icon = "🤖"
			}
			break
		}
	}

	header := fmt.Sprintf("%s %s: %s", icon, state.LastPlayedName, state.LastHandType)
	return fmt.Sprintf("%s\n%s", header, strings.Join(cardStrs, " "))
}

func renderPlayerHand(hand []card.Card, isLandlord bool) string {
	if len(hand) == 0 {
		return common.BoxStyle.Render("(无手牌)")
	}

	var rankStr, suitStr strings.Builder
	for _, c := range hand {
		style := common.BlackStyle
		if c.Color == card.Red {
			style = common.RedStyle
		}
		style = style.Align(lipgloss.Center).Margin(0, 1)
		rankStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Rank.String())))
		suitStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Suit.String())))
	}

	icon := common.FarmerIcon
	if isLandlord {
		icon = common.LandlordIcon
	}
	title := fmt.Sprintf("我的手牌 %s (%d张)", icon, len(hand))
	content := lipgloss.JoinVertical(lipgloss.Center, title, rankStr.String(), suitStr.String())
	return common.BoxStyle.Render(content)
}

func renderPrompt(m model.Model, game model.GameAccessor, state *gameClient.GameState, myPlayerID string) string {
	var sb strings.Builder
	phase := m.Phase()

	// Determine if it's player's turn
	isMyTurn := false
	switch phase {
	case model.PhaseBidding:
		isMyTurn = game.BidTurn() == myPlayerID
	case model.PhasePlaying:
		isMyTurn = state.CurrentTurn == myPlayerID
	}

	// Calculate remaining time
	timerView := renderTimer(game.TimerDuration(), game.TimerStartTime())

	switch phase {
	case model.PhaseBidding:
		action := "叫地主"
		if state.IsGrabTurn {
			action = fmt.Sprintf("抢地主 (当前倍数 ×%d)", state.Multiplier)
		}
		if game.BidTurn() == myPlayerID {
			fmt.Fprintf(&sb, "⏳ %s | 轮到你%s!\n", timerView, action)
		} else {
			for _, p := range state.Players {
				if p.ID == game.BidTurn() {
					fmt.Fprintf(&sb, "等待 %s %s...\n", p.Name, action)
					break
				}
			}
		}
	case model.PhasePlaying:
		multInfo := ""
		if state.Multiplier > 0 {
			multInfo = fmt.Sprintf(" | 💥×%d", state.Multiplier)
		}
		if state.CurrentTurn == myPlayerID {
			icon := common.FarmerIcon
			if state.IsLandlord {
				icon = common.LandlordIcon
			}
			fmt.Fprintf(&sb, "⏳ %s | 轮到你出牌! %s%s\n", timerView, icon, multInfo)
		} else {
			for _, p := range state.Players {
				if p.ID == state.CurrentTurn {
					fmt.Fprintf(&sb, "⏳ %s | 等待 %s 出牌...%s\n", timerView, p.Name, multInfo)
					break
				}
			}
		}
	}

	// Show input or quick message hint
	if isMyTurn {
		sb.WriteString(m.Input().View())
	} else {
		// When waiting, show quick message hint
		quickMsgHint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("C 键记牌器, T 键快捷消息, H 键帮助")
		sb.WriteString(quickMsgHint)
	}

	// Center the entire prompt content
	centeredContent := lipgloss.NewStyle().
		Width(m.Width()).
		AlignHorizontal(lipgloss.Center).
		Render(sb.String())

	return common.PromptStyle.Render(centeredContent)
}

func renderTimer(duration time.Duration, startTime time.Time) string {
	if startTime.IsZero() {
		return "00:00"
	}

	elapsed := time.Since(startTime)
	remaining := max(duration-elapsed, 0)

	secs := int(remaining.Seconds())
	return fmt.Sprintf("%02d:%02d", secs/60, secs%60)
}

func renderQuickMsgMenu(scroll int, inputBuf string) string {
	var sb strings.Builder
	total := len(QuickMessages)
	fmt.Fprintf(&sb, "【快捷消息】 共 %d 条\n", total)
	sb.WriteString(strings.Repeat("─", 35) + "\n")

	// Show 10 messages at a time with scroll
	visibleCount := min(10, total)
	start := scroll
	end := min(start+visibleCount, total)

	// Scroll indicator at top
	if scroll > 0 {
		sb.WriteString("    ↑ 更多消息...\n")
	}

	for i := start; i < end; i++ {
		fmt.Fprintf(&sb, "%2d. %s\n", i+1, QuickMessages[i])
	}

	// Scroll indicator at bottom
	if end < total {
		sb.WriteString("    ↓ 更多消息...\n")
	}

	sb.WriteString(strings.Repeat("─", 35) + "\n")

	// Show input prompt
	if inputBuf != "" {
		fmt.Fprintf(&sb, "输入: %s (按回车确认)\n", inputBuf)
	} else {
		sb.WriteString("输入数字选择, ↑↓滚动, T/ESC关闭\n")
	}

	return common.BoxStyle.Render(sb.String())
}

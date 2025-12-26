package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/client"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// GameModel handles game-specific logic (Waiting, Game states)
type GameModel struct {
	client *client.Client
	width  int
	height int

	input *textinput.Model

	// Game Data
	roomCode         string
	players          []protocol.PlayerInfo
	hand             []card.Card
	landlordCards    []card.Card
	currentTurn      string
	lastPlayedBy     string
	lastPlayedName   string
	lastPlayed       []card.Card
	lastHandType     string
	isLandlord       bool
	winner           string
	winnerIsLandlord bool

	// Bidding
	bidTurn string

	// State flags
	mustPlay bool
	canBeat  bool

	// Helper state
	bellPlayed     bool
	timerDuration  time.Duration
	timerStartTime time.Time

	// Features
	cardCounterEnabled bool
	remainingCards     map[card.Rank]int
	showingHelp        bool

	// Chat & Quick Messages
	chatHistory      []string
	chatInput        textinput.Model // Reuse for chat
	showQuickMsgMenu bool
}

func NewGameModel(c *client.Client, input *textinput.Model) *GameModel {
	chatInput := textinput.New()
	chatInput.Placeholder = "按 / 键聊天, T 键快捷消息..."
	chatInput.CharLimit = 50
	chatInput.Width = 30

	return &GameModel{
		client:    c,
		input:     input,
		chatInput: chatInput,
	}
}

func (m *GameModel) Init() tea.Cmd {
	return nil
}

func (m *GameModel) View() string {
	return "" // Not used directly, managed by OnlineModel
}

func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Game timer and input updates handled by parent OnlineModel for now
	// Logic can be moved here if we delegate update loop fully
	return m, nil
}

// Views

func (m *GameModel) waitingView(onlineModel *OnlineModel) string {
	var sb strings.Builder

	title := titleStyle(fmt.Sprintf("🏠 房间: %s", m.roomCode))
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	var playerList strings.Builder
	playerList.WriteString("玩家列表:\n")
	for _, p := range m.players {
		readyStr := "❌"
		if p.Ready {
			readyStr = "✅"
		}
		meStr := ""
		if p.ID == onlineModel.playerID {
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

	// Chat Rendering
	chatBox := m.renderChatBox()
	if chatBox != "" {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, chatBox))
	}

	if onlineModel.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(onlineModel.error))
		sb.WriteString(errorView)
	}

	return sb.String()
}

func (m *GameModel) gameView(onlineModel *OnlineModel) string {
	var sb strings.Builder

	topSection := m.renderTopSection()
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, topSection))
	sb.WriteString("\n")

	middleSection := m.renderMiddleSection(onlineModel.playerID)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, middleSection))
	sb.WriteString("\n")

	myHand := m.renderPlayerHand(m.hand)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, myHand))
	sb.WriteString("\n")

	prompt := m.renderPrompt(onlineModel.playerID, onlineModel.phase, &onlineModel.timer)
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, prompt))

	// Chat Rendering
	chatBox := m.renderChatBox()
	if chatBox != "" {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, chatBox))
	}

	if onlineModel.error != "" {
		errorView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "\n"+errorStyle.Render(onlineModel.error))
		sb.WriteString(errorView)
	}

	gameContent := sb.String()

	// Overlays
	if m.showQuickMsgMenu {
		menuContent := m.renderQuickMsgMenu()
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			menuContent,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	if m.showingHelp {
		helpContent := m.renderGameRules()
		helpOverlay := lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			helpContent,
			lipgloss.WithWhitespaceChars(" "),
		)
		return helpOverlay
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, gameContent)
}

func (m *GameModel) gameOverView() string {
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

func (m *GameModel) renderTopSection() string {
	landlordCardsView := m.renderLandlordCardsOnline()
	if m.cardCounterEnabled {
		cardCounter := m.renderCardCounter()
		return lipgloss.JoinHorizontal(lipgloss.Top, cardCounter, "  ", landlordCardsView)
	}
	return landlordCardsView
}

func (m *GameModel) renderLandlordCardsOnline() string {
	if len(m.landlordCards) == 0 {
		return boxStyle.Render("底牌: (待揭晓)")
	}

	var rankStr, suitStr strings.Builder
	for _, c := range m.landlordCards {
		style := blackStyle
		if c.Color == card.Red {
			style = redStyle
		}
		style = style.Align(lipgloss.Center).Margin(0, 1)
		rankStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Rank.String())))
		suitStr.WriteString(style.Render(fmt.Sprintf("%-2s", c.Suit.String())))
	}

	title := "底牌"
	content := lipgloss.JoinVertical(lipgloss.Center, title, rankStr.String(), suitStr.String())
	return boxStyle.Render(content)
}

func (m *GameModel) renderCardCounter() string {
	if !m.cardCounterEnabled {
		return ""
	}

	var sb strings.Builder
	ranks := []card.Rank{
		card.RankRedJoker, card.RankBlackJoker, card.Rank2,
		card.RankA, card.RankK, card.RankQ, card.RankJ, card.Rank10,
		card.Rank9, card.Rank8, card.Rank7, card.Rank6,
		card.Rank5, card.Rank4, card.Rank3,
	}

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

	var counts []string
	for _, rank := range ranks {
		count := m.remainingCards[rank]
		counts = append(counts, fmt.Sprintf("%-2d", count))
	}
	sb.WriteString(strings.Join(counts, "│"))

	return boxStyle.Render(sb.String())
}

func (m *GameModel) renderMiddleSection(myPlayerID string) string {
	var parts []string

	for _, p := range m.players {
		if p.ID == myPlayerID {
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

	lastPlayView := "(等待出牌...)"
	if len(m.lastPlayed) > 0 {
		var cardStrs []string
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

func (m *GameModel) renderPlayerHand(hand []card.Card) string {
	if len(hand) == 0 {
		return boxStyle.Render("(无手牌)")
	}

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

func (m *GameModel) renderPrompt(myPlayerID string, phase GamePhase, timer *timer.Model) string {
	var sb strings.Builder

	// Determine if it's player's turn
	isMyTurn := false
	switch phase {
	case PhaseBidding:
		isMyTurn = m.bidTurn == myPlayerID
	case PhasePlaying:
		isMyTurn = m.currentTurn == myPlayerID
	}

	if phase == PhaseBidding {
		if m.bidTurn == myPlayerID {
			fmt.Fprintf(&sb, "⏳ %s | 轮到你叫地主!\n", timer.View())
		} else {
			for _, p := range m.players {
				if p.ID == m.bidTurn {
					fmt.Fprintf(&sb, "等待 %s 叫地主...\n", p.Name)
					break
				}
			}
		}
	} else if phase == PhasePlaying {
		if m.currentTurn == myPlayerID {
			icon := FarmerIcon
			if m.isLandlord {
				icon = LandlordIcon
			}
			fmt.Fprintf(&sb, "⏳ %s | 轮到你出牌! %s\n", timer.View(), icon)
		} else {
			for _, p := range m.players {
				if p.ID == m.currentTurn {
					fmt.Fprintf(&sb, "等待 %s 出牌...\n", p.Name)
					break
				}
			}
		}
	}

	// Show input or quick message hint
	if isMyTurn {
		sb.WriteString(m.input.View())
	} else {
		// When waiting, show quick message hint
		quickMsgHint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("C 键记牌器, T 键快捷消息, H 键帮助")
		sb.WriteString(quickMsgHint)
	}

	return promptStyle.Render(sb.String())
}

func (m *GameModel) renderGameRules() string {
	var sb strings.Builder
	sb.WriteString("📖 斗地主游戏规则\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")

	sb.WriteString("【游戏目标】\n")
	sb.WriteString("地主：先出完手中所有牌\n")
	sb.WriteString("农民：任意一个农民先出完牌，则农民方获胜\n\n")

	sb.WriteString("【牌型说明】\n")
	sb.WriteString("• 单牌：任意一张牌\n")
	sb.WriteString("• 对子：两张点数相同的牌\n")
	sb.WriteString("• 三张：三张点数相同的牌\n")
	sb.WriteString("• 三带一：三张 + 单牌\n")
	sb.WriteString("• 三带二：三张 + 对子\n")
	sb.WriteString("• 顺子：五张或更多连续的牌（2和王不能在顺子中）\n")
	sb.WriteString("• 连对：三对或更多连续的对子\n")
	sb.WriteString("• 飞机：两个或更多连续的三张\n")
	sb.WriteString("• 四带二：四张 + 两张单牌或两个对子\n")
	sb.WriteString("• 炸弹：四张点数相同的牌（可炸任何牌型）\n")
	sb.WriteString("• 王炸：大王 + 小王（最大的牌型）\n\n")

	sb.WriteString("【叫地主规则】\n")
	sb.WriteString("1. 发牌后每位玩家依次选择是否叫地主\n")
	sb.WriteString("2. 如果有人叫地主，该玩家成为地主\n")
	sb.WriteString("3. 地主获得3张底牌，共20张牌\n")
	sb.WriteString("4. 农民各17张牌\n\n")

	sb.WriteString("【出牌规则】\n")
	sb.WriteString("1. 地主先出牌\n")
	sb.WriteString("2. 后续玩家必须出相同牌型且更大的牌，或选择PASS\n")
	sb.WriteString("3. 如果都PASS，则最后出牌的玩家可以出任意牌型\n")
	sb.WriteString("4. 炸弹和王炸可以压任何牌型\n\n")

	sb.WriteString("【快捷键】\n")
	sb.WriteString("• C：切换记牌器（游戏中）\n")
	sb.WriteString("• T：切换快捷消息（游戏中）\n")
	sb.WriteString("• H：显示/隐藏帮助（游戏中）\n")
	sb.WriteString("• ESC：返回上一级或退出\n")

	return boxStyle.Render(sb.String())
}

func (m *GameModel) rulesView() string {
	var sb strings.Builder

	title := titleStyle("📖 游戏规则")
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	sb.WriteString("\n\n")

	rules := m.renderGameRules()
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, rules))
	sb.WriteString("\n\n")

	hint := "按 ESC 返回大厅"
	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hint))

	return sb.String()
}

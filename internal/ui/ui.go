package ui

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/palemoky/fight-the-landlord-go/internal/card"
	"github.com/palemoky/fight-the-landlord-go/internal/game"
)

const (
	LandlordIcon = "👑"
	FarmerIcon   = "👨"

	TopBorderStart    = "┌──"
	TopBorderEnd      = "┌──┐"
	SideBorder        = "│"
	BottomBorderStart = "└──"
	BottomBorderEnd   = "└──┘"
)

// --- Lipgloss Styles ---
var (
	docStyle      = lipgloss.NewStyle().Margin(1, 2)
	redStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#CD0000")).Background(lipgloss.Color("#FFFFFF")).Bold(true)
	blackStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Bold(true)
	grayStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("#FFFFFF")).Bold(true)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true).Render
	boxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	promptStyle   = lipgloss.NewStyle().MarginTop(1)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	ranks         = []card.Rank{card.RankRedJoker, card.RankBlackJoker, card.Rank2, card.RankA, card.RankK, card.RankQ, card.RankJ, card.Rank10, card.Rank9, card.Rank8, card.Rank7, card.Rank6, card.Rank5, card.Rank4, card.Rank3}
)

// model 是 Bubble Tea 应用的状态
type model struct {
	game   *game.Game
	timer  timer.Model
	input  textinput.Model
	error  string
	width  int
	height int
}

// initialModel 初始化UI模型
func initialModel() model {
	g := game.NewGame()
	g.Deal()
	g.Bidding()

	ti := textinput.New()
	ti.Placeholder = "输入牌 (如 33344) 或 PASS 然后回车"
	ti.Focus()
	ti.CharLimit = 25
	ti.Width = 50

	tm := timer.NewWithInterval(game.PlayerTurnTimeout, time.Second)

	return model{
		game:  g,
		timer: tm,
		input: ti,
	}
}

func (m model) Init() tea.Cmd {
	return m.timer.Start()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			// 玩家提交出牌
			if m.game.CurrentTurn == 0 { // 确保只有轮到玩家时才能提交
				input := m.input.Value()
				m.input.Reset()
				m.error = ""

				err := m.game.PlayTurn(input)
				if err != nil {
					m.error = err.Error()
				} else {
					// 出牌成功，为下一位玩家重置计时器
					m.timer = timer.NewWithInterval(game.PlayerTurnTimeout, time.Second)
					cmds = append(cmds, m.timer.Start())
				}
			}
			return m, tea.Batch(cmds...)
		}

	case timer.TimeoutMsg:
		m.error = ""
		// 超时，自动出牌
		err := m.game.PlayTurn("") // 游戏逻辑会处理空字符串作为超时
		if err != nil {
			m.error = err.Error()
		}
		// 超时出牌更新记牌器
		// 为下一位玩家重置计时器
		m.timer = timer.NewWithInterval(game.PlayerTurnTimeout, time.Second)
		cmds = append(cmds, m.timer.Start())
	}

	m.timer, cmd = m.timer.Update(msg)
	cmds = append(cmds, cmd)

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// 游戏结束界面
	if winner, isOver := m.game.CheckWinner(); isOver {
		return m.gameOverView(winner)
	}

	// 顶部: 标题, 记牌器, 底牌
	title := titleStyle("FIGHT THE LANDLORD")
	note := "输入 Note: T->10; BJ->Black Joker; RJ->Red Joker; Pass"
	counter := m.renderCardCounter()
	landlordCards := m.renderLandlordCards()
	greetContent := lipgloss.JoinVertical(lipgloss.Center, title, note)
	counterContent := lipgloss.JoinHorizontal(lipgloss.Center, counter, landlordCards)
	topContent := lipgloss.JoinVertical(lipgloss.Center, greetContent, counterContent)
	topSection := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, topContent)

	// 中部: 其他玩家信息
	player2View := m.renderOtherPlayer(1)
	player3View := m.renderOtherPlayer(2)
	// 使用一个空的flex-box来创建间隔
	spacer := lipgloss.NewStyle().Width(m.width - 64).Render()
	middleSection := lipgloss.JoinHorizontal(lipgloss.Top, player2View, spacer, player3View)

	// 底部: 你的手牌和输入提示
	myHand := m.renderPlayerHand(m.game.Players[0].Hand)
	turnPrompt := m.renderTurnPrompt()
	bottomContent := lipgloss.JoinVertical(lipgloss.Left, myHand, turnPrompt)
	bottomSection := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, bottomContent)

	return docStyle.Render(lipgloss.JoinVertical(lipgloss.Top, topSection, middleSection, bottomSection))
}

// --- 视图渲染帮助函数 ---

func (m model) renderCard(c card.Card, content string) string {
	if c.Color == card.Red {
		return redStyle.Render(content)
	}
	return blackStyle.Render(content)
}

func (m model) renderCardCounter() string {
	var rankStr, countStr strings.Builder
	remaining := m.game.CardCounter.GetRemainingCards()

	for _, r := range ranks {
		rankStr.WriteString(fmt.Sprintf(" %-2s", r.String()))
		count := remaining[r]
		var cStr string
		if count == 0 {
			cStr = errorStyle.Render(fmt.Sprintf(" %-2d", count))
		} else {
			cStr = fmt.Sprintf(" %-2d", count)
		}
		countStr.WriteString(cStr)
	}
	content := lipgloss.JoinVertical(lipgloss.Center, "记牌器 (Card Counter)", rankStr.String(), countStr.String())
	return boxStyle.Render(content)
}

func (m model) renderLandlordCards() string {
	if len(m.game.LandlordCards) == 0 {
		return ""
	}

	var rankSB, suitSB strings.Builder
	for _, c := range m.game.LandlordCards {
		var style lipgloss.Style
		style = blackStyle
		if c.Color == card.Red {
			style = redStyle
		}
		style = style.Align(lipgloss.Center).Margin(0, 1)
		rankSB.WriteString(style.Render(fmt.Sprintf("%-2s", c.Rank.String())))
		suitSB.WriteString(style.Render(fmt.Sprintf("%-2s", c.Suit.String())))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, "底牌", rankSB.String(), suitSB.String())
	return boxStyle.Render(content)
}

func (m model) renderOtherPlayer(idx int) string {
	p := m.game.Players[idx]
	icon := FarmerIcon
	if p.IsLandlord {
		icon = LandlordIcon
	}

	nameStyle := lipgloss.NewStyle()
	if m.game.CurrentTurn == idx {
		nameStyle = nameStyle.Foreground(lipgloss.Color("220")).Bold(true)
	}
	name := nameStyle.Render(fmt.Sprintf("%s %s", icon, p.Name))
	cardsLeft := fmt.Sprintf("剩余: %d", len(p.Hand))
	var rankSB, suitSB strings.Builder
	if m.game.LastPlayerIdx == idx && !m.game.LastPlayedHand.IsEmpty() {
		for _, c := range m.game.LastPlayedHand.Cards {
			rankSB.WriteString(m.renderCard(c, c.Rank.String()) + " ")
			suitSB.WriteString(m.renderCard(c, c.Suit.String()) + " ")
		}
	}
	content := lipgloss.JoinVertical(lipgloss.Left, name, cardsLeft, "上次出牌:", rankSB.String(), suitSB.String())
	return boxStyle.Width(28).Render(content)
}

func (m model) renderFancyHand(hand []card.Card) string {
	if len(hand) == 0 {
		return "(无)"
	}

	// 我们需要为最终输出的每一行都创建一个 strings.Builder
	var top, rank, suit, bottom strings.Builder

	// 遍历除了最后一张牌之外的所有牌
	for _, c := range hand[:len(hand)-1] {
		style := blackStyle
		if c.Color == card.Red {
			style = redStyle
		}

		// 格式化点数和花色，确保'10'和'9'对齐
		rankStr := fmt.Sprintf("%-2s", c.Rank.String())
		suitStr := fmt.Sprintf("%-2s", c.Suit.String())

		// 为每一张重叠的牌只渲染左侧部分
		top.WriteString(TopBorderStart)
		rank.WriteString(SideBorder + style.Render(rankStr))
		suit.WriteString(SideBorder + style.Render(suitStr))
		bottom.WriteString(BottomBorderStart)
	}

	// 单独处理最后一张牌，渲染一个完整的、封闭的盒子
	lastCard := hand[len(hand)-1]
	style := blackStyle
	if lastCard.Color == card.Red {
		style = redStyle
	}
	rankStr := fmt.Sprintf("%-2s", lastCard.Rank.String())
	suitStr := fmt.Sprintf("%-2s", lastCard.Suit.String())

	top.WriteString(TopBorderEnd)
	rank.WriteString(SideBorder + style.Render(rankStr) + SideBorder)
	suit.WriteString(SideBorder + style.Render(suitStr) + SideBorder)
	bottom.WriteString(BottomBorderEnd)

	// 将四行拼接成最终的视图
	return lipgloss.JoinVertical(lipgloss.Left,
		top.String(),
		rank.String(),
		suit.String(),
		bottom.String(),
	)
}

func (m model) renderPlayerHand(hand []card.Card) string {
	title := "你的手牌:"
	handView := m.renderFancyHand(hand) // 调用新的渲染函数
	return lipgloss.NewStyle().MarginTop(1).Render(lipgloss.JoinVertical(lipgloss.Left, title, handView))
}

func (m model) renderTurnPrompt() string {
	currentPlayer := m.game.Players[m.game.CurrentTurn]
	var sb strings.Builder

	// 根据轮到谁来显示不同的提示和计时器
	prompt := fmt.Sprintf("⏳ %s", m.timer.View())
	if m.game.CurrentTurn == 0 { // 轮到你
		sb.WriteString(fmt.Sprintf("轮到你了, %s! %s\n", currentPlayer.Name, prompt))
		sb.WriteString(m.input.View())
		if m.error != "" {
			sb.WriteString("\n" + errorStyle.Render(m.error))
		}
	} else { // 等待其他玩家
		sb.WriteString(fmt.Sprintf("等待 %s 出牌... %s", currentPlayer.Name, prompt))
	}
	return promptStyle.Render(sb.String())
}

func (m model) gameOverView(winner *game.Player) string {
	winnerType := "农民"
	if winner.IsLandlord {
		winnerType = "地主"
	}
	msg := fmt.Sprintf("GAME OVER\n\n🥳 %s (%s) 获胜! 🎉\n\n按 Ctrl+C 或 Esc 退出", winnerType, winner.Name)
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(msg)
}

// Start 启动UI
func Start() {
	_, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run()
	if err != nil {
		log.Fatalf("启动UI时出错: %v", err)
	}
}

package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/palemoky/fight-the-landlord-go/internal/card"
	"github.com/palemoky/fight-the-landlord-go/internal/game"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

type TerminalUI struct{}

func NewTerminalUI() *TerminalUI {
	// pterm 已经处理了 reader，所以我们不再需要它
	return &TerminalUI{}
}

// renderCards 辅助函数，用于漂亮地打印一手牌
func renderCards(cards []card.Card) string {
	var sb strings.Builder
	for _, c := range cards {
		sb.WriteString(c.String())
		sb.WriteString(" ")
	}
	return sb.String()
}

// renderPlayerInfo 负责生成玩家信息区域的字符串内容
func (t *TerminalUI) renderPlayerInfo(g *game.Game) string {
	var sb strings.Builder
	for i, p := range g.Players {
		icon := "👨" // 农民图标
		style := pterm.NewStyle(pterm.FgLightWhite)
		if p.IsLandlord {
			icon = "👑" // 地主图标
			style = pterm.NewStyle(pterm.FgLightYellow, pterm.Bold)
		}
		if i == g.CurrentTurn {
			icon = "👉" + icon // 指示当前玩家
			// style = style.WithStyle(pterm.Italic)
			style = pterm.NewStyle(pterm.FgLightWhite, pterm.Italic)
		}

		sb.WriteString(style.Sprintf("%s %s", icon, p.Name))
		sb.WriteString(fmt.Sprintf("\n  剩余牌数: %d\n", len(p.Hand)))
	}
	return sb.String()
}

// renderCardCounter 负责生成记牌器表格的字符串内容
func (t *TerminalUI) renderCardCounter(g *game.Game) string {
	displayOrder := []card.Rank{
		card.RankRedJoker, card.RankBlackJoker, card.Rank2, card.RankA, card.RankK,
		card.RankQ, card.RankJ, card.Rank10, card.Rank9, card.Rank8,
		card.Rank7, card.Rank6, card.Rank5, card.Rank4, card.Rank3,
	}

	headerRow, countRow := []string{}, []string{}
	remainingCards := g.CardCounter.GetRemainingCards()

	for _, rank := range displayOrder {
		var rankCard card.Card
		if rank == card.RankRedJoker {
			rankCard = card.Card{Suit: card.Joker, Rank: rank, Color: card.Red}
		} else {
			rankCard = card.Card{Suit: card.Joker, Rank: rank, Color: card.Black}
		}
		headerRow = append(headerRow, rankCard.String())

		count := remainingCards[rank]
		var countStr string
		if count == 0 {
			countStr = pterm.NewStyle(pterm.FgRed, pterm.Strikethrough).Sprintf(" %d ", count)
		} else if count <= 2 {
			countStr = pterm.NewStyle(pterm.FgYellow).Sprintf(" %d ", count)
		} else {
			countStr = pterm.NewStyle(pterm.FgGreen).Sprintf(" %d ", count)
		}
		countRow = append(countRow, countStr)
	}

	tableData := pterm.TableData{headerRow, countRow}
	// Srender() 将组件渲染为字符串
	tableString, _ := pterm.DefaultTable.WithData(tableData).WithBoxed().Srender()
	return tableString
}

// renderGameState 负责生成场上情况区域的字符串内容
func (t *TerminalUI) renderGameState(g *game.Game) string {
	if !g.LastPlayedHand.IsEmpty() {
		lastPlayer := g.Players[g.LastPlayerIdx]
		lastPlayerName := lastPlayer.Name
		if lastPlayer.IsLandlord {
			lastPlayerName = pterm.LightYellow(lastPlayerName, " (地主)")
		}
		return fmt.Sprintf("上家 (%s) 出的牌:\n%s", lastPlayerName, renderCards(g.LastPlayedHand.Cards))
	}
	return pterm.Green("现在是你的回合, 请随意出牌。")
}

// renderPlayerHand 负责生成当前玩家手牌和提示的字符串内容
func (t *TerminalUI) renderPlayerHand(g *game.Game) {
	currentPlayer := g.Players[g.CurrentTurn]
	nameStyle := pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)
	if currentPlayer.IsLandlord {
		nameStyle = pterm.NewStyle(pterm.FgLightYellow, pterm.Bold)
	}
	pterm.DefaultSection.Printf("轮到你了, %s!", nameStyle.Sprint(currentPlayer.Name))
	pterm.Println("你的手牌:")
	pterm.Println(renderCards(currentPlayer.Hand))
	pterm.Println()
}

// DisplayGame 现在是UI布局的指挥官
func (t *TerminalUI) DisplayGame(g *game.Game) {
	t.ClearScreen()

	// 1. 渲染大标题
	logo, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("Fight The Landlord")).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.Println("Input Note: T->10; BJ->Black Joker; RJ->Red Joker; Pass")

	// 2. 获取各个区域的内容字符串
	playerInfoStr := t.renderPlayerInfo(g)
	gameStateStr := t.renderGameState(g)
	cardCounterStr := t.renderCardCounter(g)

	// 3. 使用 Columns 并排渲染“玩家信息”和“场上情况”
	// 我们将 Box 渲染成字符串 (Sprint)，然后交给 Columns 安排
	pterm.DefaultPanel.WithPanels([][]pterm.Panel{
		{
			{
				Data: pterm.DefaultBox.
					WithTitle("场上情况").
					WithTitleTopCenter().
					WithBoxStyle(pterm.NewStyle(pterm.FgLightGreen)).
					Sprint(gameStateStr),
			},
		},
		{
			{
				Data: pterm.DefaultBox.
					WithTitle("玩家信息").
					WithTitleTopCenter().
					WithBoxStyle(pterm.NewStyle(pterm.FgLightBlue)).
					Sprint(playerInfoStr),
			},
			{
				Data: pterm.DefaultBox.
					WithTitle("记牌器").
					WithTitleTopCenter().
					WithBoxStyle(pterm.NewStyle(pterm.FgLightYellow)).
					Sprint(cardCounterStr), // Println 直接渲染 Box 和其内容
			},
		},
	}).Render() // Render() 将 Columns 打印出来

	// 5. 渲染当前玩家的手牌和操作提示
	t.renderPlayerHand(g)
}

func (t *TerminalUI) GetPlayerInput(p *game.Player) string {
	// 使用 pterm 的交互式输入
	input, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultText("例如: 3334, JOKER, pass").
		Show("请出牌")

	pterm.Println()
	return strings.ToUpper(strings.TrimSpace(input))
}

func (t *TerminalUI) ShowMessage(msg string) {
	// 使用 pterm 的 Success 样式来显示通用消息
	pterm.Success.Println(msg)
	time.Sleep(2 * time.Second)
}

func (t *TerminalUI) ShowError(err error) {
	// 使用 pterm 的 Error 样式，更醒目
	pterm.Error.Println(err.Error())
	time.Sleep(2 * time.Second)
}

func (t *TerminalUI) ClearScreen() {
	cmd := exec.Command("clear") // for linux/mac
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("cmd", "/c", "cls") // for windows
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

// (可选) 新增一个游戏结束的界面
func (t *TerminalUI) DisplayGameOver(winner *game.Player, isLandlordWinner bool) {
	t.ClearScreen()
	pterm.DefaultCenter.Println(pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("GAME OVER", pterm.NewStyle(pterm.FgRed))),
	)
	pterm.Println()

	var winnerType string
	if isLandlordWinner {
		winnerType = "地主"
	} else {
		winnerType = "农民"
	}

	pterm.Success.Printf("%s (%s) 获胜!\n", winnerType, winner.Name)
	pterm.Println()
}

package ui

import (
	"bufio"
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

// renderFancyHand 负责将一手牌渲染成漂亮的、重叠的ASCII艺术风格
func (t *TerminalUI) renderFancyHand(hand []card.Card) string {
	if len(hand) == 0 {
		return "  "
	}

	// 使用三个 strings.Builder 来高效地构建每一行
	var top, rank, suit, bottom strings.Builder

	// 循环构建每一张牌的“主体”部分
	for _, c := range hand {
		rankStr := c.Rank.String()
		suitStr := c.Suit.String()

		top.WriteString("┌──")
		rank.WriteString("│" + c.RenderCard(rankStr))
		suit.WriteString("│" + c.RenderCard(suitStr))
		bottom.WriteString("└──")
	}

	// 为最后一张牌添加“封口”
	top.WriteString("┐")
	rank.WriteString("│")
	suit.WriteString("│")
	bottom.WriteString("┘")

	// 将4行拼接成最终的输出
	return fmt.Sprintf("%s\n%s\n%s\n%s", top.String(), rank.String(), suit.String(), bottom.String())
}

// renderGameState 负责生成场上情况区域的字符串内容
func (t *TerminalUI) renderGameState(g *game.Game) string {
	if !g.LastPlayedHand.IsEmpty() {
		lastPlayer := g.Players[g.LastPlayerIdx]
		lastPlayerName := lastPlayer.Name
		if lastPlayer.IsLandlord {
			lastPlayerName = pterm.LightYellow(lastPlayerName, " (地主)")
		}
		return fmt.Sprintf("上家 (%s) 出的牌:\n%s", lastPlayerName, t.renderFancyHand(g.LastPlayedHand.Cards))
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
	pterm.Println(t.renderFancyHand(currentPlayer.Hand))
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

func (t *TerminalUI) GetPlayerInput(p *game.Player, timeout time.Duration) (string, bool) {
	// 创建一个 channel 用于从 goroutine 接收输入
	inputChan := make(chan string)

	// 启动一个 goroutine 在后台等待用户输入
	// 这是一个阻塞操作，所以必须放在 goroutine 中
	go func() {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			// 如果读取出错，发送一个特殊信号或关闭 channel
			close(inputChan)
			return
		}
		inputChan <- input
	}()

	// 创建一个每秒触发一次的 Ticker
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop() // 确保函数退出时停止 ticker

	// 计算剩余时间
	remainingSeconds := int(timeout.Seconds())

	// 循环，等待输入或 ticker 触发
	for {
		// 构造并打印带倒计时的提示符
		// \r (Carriage Return) 是关键：它将光标移到行首，允许我们覆盖之前的倒计时
		prompt := pterm.LightGreen(fmt.Sprintf("\r请出牌 (剩余 %2d 秒): ", remainingSeconds))
		pterm.Print(prompt)

		select {
		case input, ok := <-inputChan:
			// 成功接收到用户输入
			if !ok {
				// Channel 被关闭，说明读取出错
				pterm.Warning.Println("\n输入读取失败！")
				return "PASS", true // 视为超时
			}
			fmt.Println()                                           // 输入完成后换行，保持界面整洁
			return strings.ToUpper(strings.TrimSpace(input)), false // 返回输入，并未超时

		case <-ticker.C:
			// Ticker 触发，时间减少一秒
			remainingSeconds--
			if remainingSeconds < 0 {
				// 倒计时结束
				pterm.Warning.Println("\n操作超时!")
				return "", true // 返回空字符串，并标记为超时
			}
		}
	}
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

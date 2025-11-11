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

const (
	Logo     = "Fight The Landlord"
	Greeting = "Input Note: T->10; BJ->Black Joker; RJ->Red Joker; Pass\n输入help或rules查看游戏规则"

	TopBorderStart    = "┌──"
	TopBorderEnd      = "┐"
	SideBorder        = "│"
	SeparatorStart    = "├──"
	SeparatorEnd      = "┤"
	BottomBorderStart = "└──"
	BottomBorderEnd   = "┘"
)

type TerminalUI struct{}

func NewTerminalUI() *TerminalUI {
	// pterm 已经处理了 reader，所以我们不再需要它
	return &TerminalUI{}
}

func renderCards(cards []card.Card) string {
	var sb strings.Builder
	for _, c := range cards {
		sb.WriteString(c.String())
		sb.WriteString(" ")
	}
	return sb.String()
}

// renderCardContent 将单张牌渲染成带样式的字符串内容，例如 "♥A"
func (t *TerminalUI) renderCardContent(c card.Card, str string) string {
	styleRed := pterm.NewRGBStyle(pterm.NewRGB(192, 0, 0), pterm.NewRGB(228, 215, 215))
	styleBlack := pterm.NewRGBStyle(pterm.NewRGB(68, 67, 77), pterm.NewRGB(228, 215, 215))

	content := fmt.Sprintf("%-2s", str)
	// content := t.getCardContentString(c)

	styledCard := styleBlack.Sprint(content)
	if c.Color == card.Red {
		styledCard = styleRed.Sprint(content)
	}

	return styledCard
}

// getCardContentString 获取卡牌内容的纯字符串，用于统一格式
func (t *TerminalUI) getCardContentString(c card.Card) string {
	suitStr := c.Suit.String()
	if c.Suit == card.Joker {
		suitStr = ""
	}
	return fmt.Sprintf("%-5s", suitStr+c.Rank.String())
}

// renderFancyHand 负责将一手牌渲染成漂亮的、重叠的ASCII艺术风格
func (t *TerminalUI) renderFancyHand(hand []card.Card, g *game.Game) string {
	if len(hand) == 0 {
		return pterm.Gray(" ")
	}

	// 灰字 + 浅灰背景
	stylePlayed := pterm.NewRGBStyle(pterm.NewRGB(128, 128, 128), pterm.NewRGB(220, 220, 220))

	var top, rank, suit, bottom strings.Builder
	for _, c := range hand {
		isPlayed := false
		// 如果传入了 game 对象 (即我们正在渲染底牌), 则检查卡牌状态
		if g != nil {
			isPlayed = g.IsLandlordCardPlayed(c)
		}

		top.WriteString(TopBorderStart)
		if isPlayed {
			// 如果牌已出，使用“已出”样式
			// styledContent = stylePlayed.Sprint(t.getCardContentString(c))
			rank.WriteString(SideBorder + stylePlayed.Sprint(t.getCardContentString(c)))
			suit.WriteString(SideBorder + stylePlayed.Sprint(t.getCardContentString(c)))
		} else {
			rank.WriteString(SideBorder + t.renderCardContent(c, c.Rank.String()))
			suit.WriteString(SideBorder + t.renderCardContent(c, c.Suit.String()))
		}
		bottom.WriteString(BottomBorderStart)
	}

	top.WriteString(TopBorderEnd)
	rank.WriteString(SideBorder)
	suit.WriteString(SideBorder)
	bottom.WriteString(BottomBorderEnd)

	return fmt.Sprintf("%s\n%s\n%s\n%s", top.String(), rank.String(), suit.String(), bottom.String())
}

// renderPlayerInfo 负责生成玩家信息区域的字符串内容
func (t *TerminalUI) renderPlayerInfoBox(g *game.Game) string {
	var sb strings.Builder
	for i, p := range g.Players {
		icon := "👨" // farmer icon
		style := pterm.NewStyle(pterm.FgLightWhite)
		if p.IsLandlord {
			icon = "👑" // landlord icon
			style = pterm.NewStyle(pterm.FgLightYellow, pterm.Bold)
		}
		if i == g.CurrentTurn {
			icon = "👉" + icon // current player
			style = pterm.NewStyle(pterm.FgLightWhite, pterm.Italic)
		}

		sb.WriteString(style.Sprintf("%s %s", icon, p.Name))
		sb.WriteString(fmt.Sprintf("\n  剩余: %d\n", len(p.Hand)))

		// 显示上次出牌
		sb.WriteString("上次出牌: ")
		if i == g.LastPlayerIdx && !g.LastPlayedHand.IsEmpty() {
			// 只为上一个出牌的玩家显示其出的牌
			sb.WriteString("\n")
			// 使用简单的 renderCards 避免占用太多空间
			sb.WriteString(renderCards(g.LastPlayedHand.Cards))
		} else {
			sb.WriteString(pterm.Gray("(无)"))
		}
		sb.WriteString("\n\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// renderPlayerHand 负责生成当前玩家手牌和提示的字符串内容
func (t *TerminalUI) renderPlayerHand(g *game.Game) {
	currentPlayer := g.Players[g.CurrentTurn]
	nameStyle := pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)
	if currentPlayer.IsLandlord {
		nameStyle = pterm.NewStyle(pterm.FgLightYellow, pterm.Bold)
	}
	pterm.DefaultSection.Printf("轮到你了, %s!", nameStyle.Sprint(currentPlayer.Name))
	// pterm.Println("你的手牌:")
	pterm.Println(t.renderFancyHand(currentPlayer.Hand, nil))
	pterm.Println()
}

// renderCounterGrid 手动绘制记牌器
func (t *TerminalUI) renderCounterGrid(g *game.Game) string {
	displayOrder := []card.Rank{
		card.RankRedJoker, card.RankBlackJoker, card.Rank2, card.RankA, card.RankK,
		card.RankQ, card.RankJ, card.Rank10, card.Rank9, card.Rank8,
		card.Rank7, card.Rank6, card.Rank5, card.Rank4, card.Rank3,
	}

	var top, ranks, cards, separator, counts, bottom strings.Builder
	remainingCards := g.CardCounter.GetRemainingCards()

	for _, rank := range displayOrder {
		// --- 构建牌面行 ---
		rankCard := card.Card{Suit: card.Joker, Rank: rank, Color: card.Black}
		if rank == card.RankRedJoker {
			rankCard = card.Card{Suit: card.Joker, Rank: rank, Color: card.Red}
		}
		// 复用 renderCardContent 来获取带样式的牌面内容
		ranks.WriteString(SideBorder + t.renderCardContent(rankCard, rank.String()))
		cards.WriteString(SideBorder + t.renderCardContent(rankCard, " "))

		// --- 构建数量行 ---
		count := remainingCards[rank]
		var countStr string
		if count == 0 {
			countStr = pterm.NewStyle(pterm.FgRed, pterm.Strikethrough).Sprintf("%d ", count)
		} else if count <= 2 {
			countStr = pterm.NewStyle(pterm.FgYellow).Sprintf("%d ", count)
		} else {
			countStr = pterm.NewStyle(pterm.FgGreen).Sprintf("%d ", count)
		}
		counts.WriteString(SideBorder + countStr)

		top.WriteString(TopBorderStart)
		separator.WriteString(SeparatorStart)
		bottom.WriteString(BottomBorderStart)
	}

	top.WriteString(TopBorderEnd)
	ranks.WriteString(SideBorder)
	cards.WriteString(SideBorder)
	separator.WriteString(SeparatorEnd)
	counts.WriteString(SideBorder)
	bottom.WriteString(BottomBorderEnd)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		top.String(), ranks.String(), cards.String(), separator.String(), counts.String(), bottom.String())
}

// DisplayGame 总指挥
func (t *TerminalUI) DisplayGame(g *game.Game) {
	t.ClearScreen()

	// 1. 渲染大标题
	logo, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString(Logo)).Srender()
	pterm.DefaultCenter.Println(logo)
	pterm.DefaultCenter.Println(Greeting)

	playerInfoContent := t.renderPlayerInfoBox(g) // 玩家信息
	counterGridStr := t.renderCounterGrid(g)      // 记牌器

	// 底牌信息
	var landlordCardsBuilder strings.Builder
	landlordCardsBuilder.WriteString(t.renderFancyHand(g.LandlordCards, g))
	landlordCardsStr := landlordCardsBuilder.String()

	paddedBox := pterm.DefaultBox
	playerInfo := paddedBox.WithTitle("玩家信息 (Player Info)").Sprint(playerInfoContent)
	cardCounter := paddedBox.WithTitle("记牌器 (Card Counter)").Sprint(counterGridStr)
	landlordsCards := paddedBox.WithTitle("底牌").WithTitleTopCenter().Sprint(landlordCardsStr)
	pterm.DefaultPanel.WithPanels([][]pterm.Panel{
		{{Data: cardCounter}},
		{{Data: playerInfo}, {Data: landlordsCards}},
	}).Render()

	// 渲染当前玩家的手牌和操作提示
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

// 新增一个游戏结束的界面
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

	pterm.Success.Printf("🥳 %s (%s) 获胜! 🎉\n", winnerType, winner.Name)
	pterm.Println()
}

func (t *TerminalUI) DisplayRules() {
	t.ClearScreen()

	// 使用 Header 制作一个漂亮的标题
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).Println("游戏规则 (Game Rules)")
	pterm.Println()

	// 使用 BulletList 来格式化规则列表，非常清晰
	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: "单张 (Single): 任意一张牌。"},
		{Level: 0, Text: "对子 (Pair): 两张点数相同的牌。"},
		{Level: 0, Text: "三张 (Trio): 三张点数相同的牌。"},
		{Level: 1, Text: "三带一 (Trio with Single): 三张 + 一张单牌。"},
		{Level: 1, Text: "三带二 (Trio with Pair): 三张 + 一个对子。"},
		{Level: 0, Text: "顺子 (Straight): 5张或以上点数连续的单牌 (A, 2, 王除外)。"},
		{Level: 0, Text: "连对 (Pair Straight): 3对或以上点数连续的对子 (A, 2, 王除外)。"},
		{Level: 0, Text: "飞机 (Plane): 2个或以上点数连续的三张 (A, 2, 王除外)。"},
		{Level: 1, Text: "飞机带单 (Plane with Singles): 飞机 + 对应数量的单牌。"},
		{Level: 1, Text: "飞机带对 (Plane with Pairs): 飞机 + 对应数量的对子。"},
		{Level: 0, Text: "炸弹 (Bomb): 四张点数相同的牌。"},
		{Level: 1, Text: "四带二 (Four with Two): 四张 + 两张单牌或一个对子。"},
		{Level: 0, Text: "王炸 (Rocket): 红Joker + 黑Joker，最大的牌型。"},
	}).Render()

	pterm.Println()

	// 交互式提示，等待用户按键后返回游戏
	pterm.DefaultInteractiveContinue.Show("按回车键返回游戏...")
}

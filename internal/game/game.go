package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/palemoky/fight-the-landlord-go/internal/card"
	"github.com/palemoky/fight-the-landlord-go/internal/rule"
)

// Game 定义游戏状态
type Game struct {
	Players           [3]*Player
	Deck              card.Deck
	LandlordCards     []card.Card
	CurrentTurn       int
	LastPlayedHand    rule.ParsedHand
	LastPlayerIdx     int
	ConsecutivePasses int
	CardCounter       *card.CardCounter
	ui                UI
}

// UI 是一个接口，定义了游戏与用户交互的所有方法
type UI interface {
	DisplayGame(*Game)
	GetPlayerInput(*Player) string
	ShowMessage(string)
	ShowError(error)
	ClearScreen()
}

func NewGame(ui UI) *Game {
	players := [3]*Player{
		{Name: "Player 1"},
		{Name: "Player 2"},
		{Name: "Player 3"},
	}
	deck := card.NewDeck()
	deck.Shuffle()

	return &Game{
		Players:     players,
		Deck:        deck,
		CardCounter: card.NewCardCounter(),
		ui:          ui,
	}
}

func (g *Game) Deal() {
	for range 17 {
		for _, p := range g.Players {
			p.Hand = append(p.Hand, g.Deck[0])
			g.Deck = g.Deck[1:]
		}
	}
	g.LandlordCards = g.Deck
	for _, p := range g.Players {
		p.SortHand()
	}
}

func (g *Game) Bidding() {
	g.ui.ClearScreen()
	g.ui.ShowMessage("--- 叫地主阶段 ---")
	bidderIdx := rand.Intn(3)
	fmt.Printf("从 %s 开始叫地主。\n", g.Players[bidderIdx].Name)

	time.Sleep(2 * time.Second)

	g.Players[bidderIdx].IsLandlord = true
	g.Players[bidderIdx].Hand = append(g.Players[bidderIdx].Hand, g.LandlordCards...)
	g.Players[bidderIdx].SortHand()

	g.CurrentTurn = bidderIdx
	g.LastPlayerIdx = bidderIdx

	g.ui.ShowMessage(fmt.Sprintf("%s 成为地主！并获得三张底牌。\n", g.Players[bidderIdx].Name))
	g.ui.ShowMessage(fmt.Sprintf("底牌是: %s", g.LandlordCards))
	time.Sleep(3 * time.Second)
}

func (g *Game) Run() {
	g.Deal()
	g.Bidding()

	for {
		g.ui.DisplayGame(g)
		currentPlayer := g.Players[g.CurrentTurn]

		input := g.ui.GetPlayerInput(currentPlayer)

		if input == "PASS" {
			if g.LastPlayerIdx == g.CurrentTurn {
				fmt.Println("你是第一个出牌的，不能 PASS！")
				time.Sleep(2 * time.Second)
				continue
			}
			g.ConsecutivePasses++
			if g.ConsecutivePasses == 2 {
				fmt.Println("所有人都 PASS，轮到上一轮出牌者重新出牌。")
				g.LastPlayedHand = rule.ParsedHand{} // 使用 rule.ParsedHand
				g.ConsecutivePasses = 0
				g.LastPlayerIdx = g.CurrentTurn // 轮到上个出牌者继续出牌
			}
			g.CurrentTurn = (g.CurrentTurn + 1) % 3
			continue
		}

		cardsToPlay, err := rule.FindCardsInHand(currentPlayer.Hand, input) 
		if err != nil {
			fmt.Println("出牌无效: ", err)
			time.Sleep(2 * time.Second)
			continue
		}

		handToPlay, err := rule.ParseHand(cardsToPlay)
		if err != nil {
			fmt.Println("无效的牌型: ", err)
			time.Sleep(2 * time.Second)
			continue
		}

		canPlay := false
		if g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() {
			canPlay = true
		} else {
			canPlay = rule.CanBeat(handToPlay, g.LastPlayedHand)
		}

		if canPlay {
			g.LastPlayedHand = handToPlay
			g.LastPlayerIdx = g.CurrentTurn
			g.ConsecutivePasses = 0

			g.CardCounter.Update(cardsToPlay) // 出牌后更新记牌器
			currentPlayer.Hand = rule.RemoveCards(currentPlayer.Hand, cardsToPlay)

			if len(currentPlayer.Hand) == 0 {

				g.ui.DisplayGame(g)
				fmt.Println("\n================== 游戏结束 ==================")
				if currentPlayer.IsLandlord {
					fmt.Printf("地主 (%s) 获胜!\n", currentPlayer.Name)
				} else {
					fmt.Printf("农民 (%s) 获胜!\n", currentPlayer.Name)
				}
				return
			}
			g.CurrentTurn = (g.CurrentTurn + 1) % 3
		} else {
			fmt.Println("你的牌没有大过上家!")
			time.Sleep(2 * time.Second)
		}
	}
}

package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/palemoky/fight-the-landlord-go/internal/card"
	"github.com/palemoky/fight-the-landlord-go/internal/rule"
)

const (
	PlayerTurnTimeout = 30 * time.Second
)

// Game 定义游戏状态
type Game struct {
	Players              [3]*Player
	Deck                 card.Deck
	LandlordCards        []card.Card
	CurrentTurn          int
	LastPlayedHand       rule.ParsedHand
	LastPlayerIdx        int
	ConsecutivePasses    int
	CardCounter          *card.CardCounter
	CanCurrentPlayerPlay bool
}

// NewGame 初始化一个新游戏
func NewGame() *Game {
	players := [3]*Player{
		{Name: "Player 1 (你)"},
		{Name: "Player 2"},
		{Name: "Player 3"},
	}
	deck := card.NewDeck()
	deck.Shuffle()

	return &Game{
		Players:              players,
		Deck:                 deck,
		CardCounter:          card.NewCardCounter(),
		CanCurrentPlayerPlay: true, // 游戏开始时，第一个玩家总是有牌可出
	}
}

// Deal 发牌
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

// Bidding 叫地主（此处为简化版，随机选择一个）
func (g *Game) Bidding() {
	landlordIdx := rand.Intn(3)
	g.Players[landlordIdx].IsLandlord = true
	g.Players[landlordIdx].Hand = append(g.Players[landlordIdx].Hand, g.LandlordCards...)
	g.Players[landlordIdx].SortHand()

	g.CurrentTurn = landlordIdx
	g.LastPlayerIdx = landlordIdx
}

// PlayTurn 处理玩家的一次出牌操作
func (g *Game) PlayTurn(input string) error {
	currentPlayer := g.Players[g.CurrentTurn]

	// 1. 预处理输入，处理超时情况
	processedInput := g.preprocessInput(input, currentPlayer)

	// 2. 根据输入，分派给不同的处理函数
	var err error
	if strings.ToUpper(strings.TrimSpace(processedInput)) == "PASS" {
		err = g.handlePass()
	} else {
		err = g.handlePlay(currentPlayer, processedInput)
	}

	// 3. 如果处理过程中出现错误，立即返回
	if err != nil {
		return err
	}

	// 4. 如果游戏已经结束，直接返回
	if len(currentPlayer.Hand) == 0 {
		g.CanCurrentPlayerPlay = false // 游戏结束，下一个玩家无所谓了
		return nil
	}

	// 5. 如果回合成功，则推进到下一回合，并更新状态
	g.advanceToNextTurn()

	return nil
}

// preprocessInput 负责处理超时逻辑，返回一个确定的指令 ("PASS" 或出牌字符串)
func (g *Game) preprocessInput(input string, currentPlayer *Player) string {
	isTimeout := input == ""
	if !isTimeout {
		return input
	}

	// 如果是轮到你自由出牌，不能pass，自动打出最小的单牌
	if g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() {
		minCard := currentPlayer.Hand[len(currentPlayer.Hand)-1] // 手牌已排序
		return minCard.Rank.String()
	}

	// 否则自动PASS
	return "PASS"
}

// handlePass 专门处理玩家选择 PASS 的逻辑
func (g *Game) handlePass() error {
	if g.LastPlayerIdx == g.CurrentTurn || g.ConsecutivePasses == 2 {
		return errors.New("轮到你出牌，不能PASS")
	}
	g.ConsecutivePasses++
	if g.ConsecutivePasses == 2 {
		// 如果连续两人PASS，则开启新的一轮
		g.LastPlayedHand = rule.ParsedHand{}
		g.LastPlayerIdx = (g.CurrentTurn + 1) % 3 // 新一轮由下家开始
	}
	return nil
}

// handlePlay 专门处理玩家出牌的逻辑
func (g *Game) handlePlay(currentPlayer *Player, input string) error {
	cardsToPlay, err := card.FindCardsInHand(currentPlayer.Hand, strings.ToUpper(input))
	if err != nil {
		return fmt.Errorf("出牌无效: %w", err)
	}

	handToPlay, err := rule.ParseHand(cardsToPlay)
	if err != nil {
		return fmt.Errorf("无效的牌型: %w", err)
	}

	isNewRound := g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() || g.ConsecutivePasses == 2
	if isNewRound || rule.CanBeat(handToPlay, g.LastPlayedHand) {
		// 出牌成功，更新游戏状态
		g.LastPlayedHand = handToPlay
		g.LastPlayerIdx = g.CurrentTurn
		g.ConsecutivePasses = 0
		g.CardCounter.Update(cardsToPlay)
		currentPlayer.Hand = card.RemoveCards(currentPlayer.Hand, cardsToPlay)

		return nil
	} else {
		return errors.New("你的牌没有大过上家")
	}
}

// advanceToNextTurn 推进回合，并为下一个玩家设置状态
func (g *Game) advanceToNextTurn() {
	// 1. 将回合交给下一个玩家
	g.CurrentTurn = (g.CurrentTurn + 1) % 3

	// 2. 获取下一个玩家
	nextPlayer := g.Players[g.CurrentTurn]

	// 3. 判断下一个玩家是否可以自由出牌
	isFreePlay := g.LastPlayedHand.IsEmpty() || g.LastPlayerIdx == g.CurrentTurn
	if isFreePlay {
		g.CanCurrentPlayerPlay = true
	} else {
		// 否则，检查他是否有牌可打
		g.CanCurrentPlayerPlay = rule.CanBeatWithHand(nextPlayer.Hand, g.LastPlayedHand)
	}
}

// CheckWinner 检查是否有玩家获胜
func (g *Game) CheckWinner() (*Player, bool) {
	for _, p := range g.Players {
		if len(p.Hand) == 0 {
			return p, true
		}
	}
	return nil, false
}

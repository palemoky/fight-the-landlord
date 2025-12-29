package game

import (
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/rule"
)

// HandlePlayCards 处理出牌
func (gs *GameSession) HandlePlayCards(playerID string, cardInfos []protocol.CardInfo) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.state != GameStatePlaying {
		return ErrGameNotStart
	}

	currentPlayer := gs.players[gs.currentPlayer]
	if currentPlayer.ID != playerID {
		return ErrNotYourTurn
	}

	// 取消超时计时器
	gs.stopTimer()

	// 转换牌
	cards := convert.InfosToCards(cardInfos)

	// 验证牌是否在手中
	if !gs.validateCardsInHand(currentPlayer, cards) {
		return ErrInvalidCards
	}

	// 解析牌型
	handToPlay, err := rule.ParseHand(cards)
	if err != nil {
		return ErrInvalidCards
	}

	// 检查是否能打过上家
	isNewRound := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()
	if !isNewRound && !rule.CanBeat(handToPlay, gs.lastPlayedHand) {
		return ErrCannotBeat
	}

	// 出牌成功，更新状态
	gs.lastPlayedHand = handToPlay
	gs.lastPlayerIdx = gs.currentPlayer
	gs.consecutivePasses = 0

	// 从手牌中移除
	currentPlayer.Hand = card.RemoveCards(currentPlayer.Hand, cards)

	// 对出的牌进行排序（从大到小），确保显示顺序正确
	sortedCards := make([]card.Card, len(cards))
	copy(sortedCards, cards)
	sort.Slice(sortedCards, func(i, j int) bool {
		return sortedCards[i].Rank > sortedCards[j].Rank
	})

	// 广播出牌信息
	gs.room.broadcast(encoding.MustNewMessage(protocol.MsgCardPlayed, protocol.CardPlayedPayload{
		PlayerID:   playerID,
		PlayerName: currentPlayer.Name,
		Cards:      convert.CardsToInfos(sortedCards), // 使用排序后的牌
		CardsLeft:  len(currentPlayer.Hand),
		HandType:   handToPlay.Type.String(),
	}))

	// 检查是否获胜
	if len(currentPlayer.Hand) == 0 {
		gs.endGame(currentPlayer)
		return nil
	}

	// 下一个玩家
	gs.currentPlayer = (gs.currentPlayer + 1) % 3
	gs.notifyPlayTurn()

	return nil
}

// HandlePass 处理不出
func (gs *GameSession) HandlePass(playerID string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.state != GameStatePlaying {
		return ErrGameNotStart
	}

	currentPlayer := gs.players[gs.currentPlayer]
	if currentPlayer.ID != playerID {
		return ErrNotYourTurn
	}

	// 检查是否必须出牌
	mustPlay := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()
	if mustPlay {
		return ErrMustPlay
	}

	// 取消超时计时器
	gs.stopTimer()

	gs.consecutivePasses++

	// 广播不出
	gs.room.broadcast(encoding.MustNewMessage(protocol.MsgPlayerPass, protocol.PlayerPassPayload{
		PlayerID:   playerID,
		PlayerName: currentPlayer.Name,
	}))

	// 如果连续两人不出，新一轮开始
	if gs.consecutivePasses >= 2 {
		gs.lastPlayedHand = rule.ParsedHand{}
		gs.lastPlayerIdx = (gs.currentPlayer + 1) % 3
		gs.consecutivePasses = 0
	}

	// 下一个玩家
	gs.currentPlayer = (gs.currentPlayer + 1) % 3
	gs.notifyPlayTurn()

	return nil
}

// validateCardsInHand 验证牌是否在手中
func (gs *GameSession) validateCardsInHand(player *GamePlayer, cards []card.Card) bool {
	handCopy := make([]card.Card, len(player.Hand))
	copy(handCopy, player.Hand)

	for _, c := range cards {
		found := false
		for i, h := range handCopy {
			if h.Suit == c.Suit && h.Rank == c.Rank {
				handCopy = append(handCopy[:i], handCopy[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GetPlayerCardsCount 获取玩家手牌数量
func (gs *GameSession) GetPlayerCardsCount(playerID string) int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	for _, p := range gs.players {
		if p.ID == playerID {
			return len(p.Hand)
		}
	}
	return 0
}

// notifyPlayTurn 通知当前玩家出牌
func (gs *GameSession) notifyPlayTurn() {
	player := gs.players[gs.currentPlayer]
	mustPlay := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()
	canBeat := true // 简化处理

	gs.room.Broadcast(encoding.MustNewMessage(protocol.MsgPlayTurn, protocol.PlayTurnPayload{
		PlayerID: player.ID,
		Timeout:  30,
		MustPlay: mustPlay,
		CanBeat:  canBeat,
	}))
	gs.startPlayTimer()
}

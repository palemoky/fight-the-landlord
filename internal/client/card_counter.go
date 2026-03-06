package client

import "github.com/palemoky/fight-the-landlord/internal/game/card"

// CardCounter 跟踪不在玩家手中的剩余牌
type CardCounter struct {
	remaining map[card.Rank]int
}

// NewCardCounter 创建并初始化一个新的记牌器
func NewCardCounter() *CardCounter {
	cc := &CardCounter{
		remaining: make(map[card.Rank]int),
	}
	cc.Reset()
	return cc
}

// Reset 使用完整的一副牌（54张）初始化计数器
func (cc *CardCounter) Reset() {
	// 3-A 和 2 各有 4 张牌
	for rank := card.Rank3; rank <= card.Rank2; rank++ {
		cc.remaining[rank] = 4
	}

	// 王各有 1 张
	cc.remaining[card.RankBlackJoker] = 1
	cc.remaining[card.RankRedJoker] = 1
}

// DeductCards 从计数器中扣除已出的牌
func (cc *CardCounter) DeductCards(cards []card.Card) {
	for _, c := range cards {
		if cc.remaining[c.Rank] > 0 {
			cc.remaining[c.Rank]--
		}
	}
}

// GetRemaining 返回剩余牌的计数
func (cc *CardCounter) GetRemaining() map[card.Rank]int {
	return cc.remaining
}

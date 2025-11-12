package card

// CardCounter 记牌器，记录场上还剩下哪些牌
type CardCounter struct {
	remainingCards map[Rank]int
}

// NewCardCounter 创建并初始化一个记牌器
func NewCardCounter() *CardCounter {
	counter := &CardCounter{
		remainingCards: make(map[Rank]int, 15),
	}

	// 初始化一副完整的牌
	// 每种点数 (3-2) 都有4张
	for r := Rank3; r <= Rank2; r++ {
		counter.remainingCards[r] = 4
	}
	// 大小王各一张
	counter.remainingCards[RankBlackJoker] = 1
	counter.remainingCards[RankRedJoker] = 1

	return counter
}

// Update 根据出掉的牌来更新记牌器
func (cc *CardCounter) Update(playedCards []Card) {
	for _, c := range playedCards {
		if _, ok := cc.remainingCards[c.Rank]; ok {
			cc.remainingCards[c.Rank]--
		}
	}
}

func (cc *CardCounter) GetRemainingCards() map[Rank]int {
	return cc.remainingCards
}

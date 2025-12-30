package client

import "github.com/palemoky/fight-the-landlord/internal/card"

// CardCounter tracks remaining cards not in player's hand
type CardCounter struct {
	remaining map[card.Rank]int
}

// NewCardCounter creates and initializes a new card counter
func NewCardCounter() *CardCounter {
	cc := &CardCounter{
		remaining: make(map[card.Rank]int),
	}
	cc.Reset()
	return cc
}

// Reset initializes counter with a full deck (54 cards)
func (cc *CardCounter) Reset() {
	// 3-A and 2 each have 4 cards
	for rank := card.Rank3; rank <= card.RankA; rank++ {
		cc.remaining[rank] = 4
	}
	cc.remaining[card.Rank2] = 4

	// Jokers have 1 each
	cc.remaining[card.RankBlackJoker] = 1
	cc.remaining[card.RankRedJoker] = 1
}

// DeductCards removes cards from the counter
func (cc *CardCounter) DeductCards(cards []card.Card) {
	for _, c := range cards {
		if cc.remaining[c.Rank] > 0 {
			cc.remaining[c.Rank]--
		}
	}
}

// GetRemaining returns the remaining card counts
func (cc *CardCounter) GetRemaining() map[card.Rank]int {
	return cc.remaining
}

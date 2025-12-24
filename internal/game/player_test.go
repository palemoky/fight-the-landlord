package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/card"
)

func testPlayerCards(ranks ...card.Rank) []card.Card {
	cards := make([]card.Card, len(ranks))
	for i, r := range ranks {
		cards[i] = card.Card{Rank: r}
	}
	return cards
}

func TestPlayer_SortHand(t *testing.T) {
	testCases := []struct {
		name         string
		initialHand  []card.Card
		expectedHand []card.Card
	}{
		{
			name:         "standard unsorted hand",
			initialHand:  testPlayerCards(card.Rank5, card.RankA, card.Rank3, card.RankK),
			expectedHand: testPlayerCards(card.RankA, card.RankK, card.Rank5, card.Rank3),
		},
		{
			name:         "hand with jokers",
			initialHand:  testPlayerCards(card.RankK, card.RankBlackJoker, card.RankA, card.RankRedJoker, card.Rank2),
			expectedHand: testPlayerCards(card.RankRedJoker, card.RankBlackJoker, card.Rank2, card.RankA, card.RankK),
		},
		{
			name:         "hand with duplicate ranks",
			initialHand:  testPlayerCards(card.Rank4, card.Rank8, card.Rank4, card.RankJ, card.Rank8, card.Rank8),
			expectedHand: testPlayerCards(card.RankJ, card.Rank8, card.Rank8, card.Rank8, card.Rank4, card.Rank4),
		},
		{
			name:         "already sorted hand",
			initialHand:  testPlayerCards(card.RankQ, card.Rank10, card.Rank8),
			expectedHand: testPlayerCards(card.RankQ, card.Rank10, card.Rank8),
		},
		{
			name:         "hand with a single card",
			initialHand:  testPlayerCards(card.Rank7),
			expectedHand: testPlayerCards(card.Rank7),
		},
		{
			name:         "empty hand",
			initialHand:  []card.Card{},
			expectedHand: []card.Card{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			player := Player{Hand: tc.initialHand}
			player.SortHand()
			assert.Equal(t, tc.expectedHand, player.Hand, "The hand should be sorted correctly.")
		})
	}
}

package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/card"
)

func TestCardConversion(t *testing.T) {
	tests := []struct {
		name string
		c    card.Card
	}{
		{"Ace of Spades", card.Card{Suit: card.Spade, Rank: card.RankA, Color: card.Black}},
		{"3 of Hearts", card.Card{Suit: card.Heart, Rank: card.Rank3, Color: card.Red}},
		{"Small Joker", card.Card{Suit: card.Joker, Rank: card.RankBlackJoker, Color: card.Black}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Card -> Info
			info := CardToInfo(tt.c)
			assert.Equal(t, int(tt.c.Suit), info.Suit)
			assert.Equal(t, int(tt.c.Rank), info.Rank)
			assert.Equal(t, int(tt.c.Color), info.Color)

			// Info -> Card
			convertedCard := InfoToCard(info)
			assert.Equal(t, tt.c, convertedCard)
		})
	}
}

func TestCardsConversion(t *testing.T) {
	cards := []card.Card{
		{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
		{Suit: card.Heart, Rank: card.Rank4, Color: card.Red},
	}

	// Cards -> Infos
	infos := CardsToInfos(cards)
	assert.Len(t, infos, 2)

	// Infos -> Cards
	convertedCards := InfosToCards(infos)
	assert.Equal(t, cards, convertedCards)
}

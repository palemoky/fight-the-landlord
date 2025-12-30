package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

func TestCardToInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		card     card.Card
		expected protocol.CardInfo
	}{
		{
			name: "Ace of Spades",
			card: card.Card{
				Suit:  card.Spade,
				Rank:  card.RankA,
				Color: card.Black,
			},
			expected: protocol.CardInfo{
				Suit:  int(card.Spade),
				Rank:  int(card.RankA),
				Color: int(card.Black),
			},
		},
		{
			name: "Red Joker",
			card: card.Card{
				Suit:  card.Joker,
				Rank:  card.RankRedJoker,
				Color: card.Red,
			},
			expected: protocol.CardInfo{
				Suit:  int(card.Joker),
				Rank:  int(card.RankRedJoker),
				Color: int(card.Red),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := CardToInfo(tt.card)
			assert.Equal(t, tt.expected, info)
		})
	}
}

func TestInfoToCard(t *testing.T) {
	t.Parallel()

	info := protocol.CardInfo{
		Suit:  int(card.Heart),
		Rank:  int(card.RankK),
		Color: int(card.Red),
	}

	result := InfoToCard(info)

	assert.Equal(t, card.Heart, result.Suit)
	assert.Equal(t, card.RankK, result.Rank)
	assert.Equal(t, card.Red, result.Color)
}

func TestCardsToInfos(t *testing.T) {
	t.Parallel()

	cards := []card.Card{
		{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
		{Suit: card.Heart, Rank: card.RankA, Color: card.Red},
	}

	infos := CardsToInfos(cards)

	require.Len(t, infos, 2)
	assert.Equal(t, int(card.Spade), infos[0].Suit)
	assert.Equal(t, int(card.Rank3), infos[0].Rank)
	assert.Equal(t, int(card.Heart), infos[1].Suit)
	assert.Equal(t, int(card.RankA), infos[1].Rank)
}

func TestInfosToCards(t *testing.T) {
	t.Parallel()

	infos := []protocol.CardInfo{
		{Suit: int(card.Diamond), Rank: int(card.Rank5), Color: int(card.Red)},
		{Suit: int(card.Club), Rank: int(card.Rank10), Color: int(card.Black)},
	}

	cards := InfosToCards(infos)

	require.Len(t, cards, 2)
	assert.Equal(t, card.Diamond, cards[0].Suit)
	assert.Equal(t, card.Rank5, cards[0].Rank)
	assert.Equal(t, card.Club, cards[1].Suit)
	assert.Equal(t, card.Rank10, cards[1].Rank)
}

func TestCardRoundTrip(t *testing.T) {
	t.Parallel()

	original := card.Card{
		Suit:  card.Spade,
		Rank:  card.Rank2,
		Color: card.Black,
	}

	// Card -> Info -> Card
	info := CardToInfo(original)
	result := InfoToCard(info)

	assert.Equal(t, original, result)
}

func TestCardsRoundTrip(t *testing.T) {
	t.Parallel()

	originals := []card.Card{
		{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
		{Suit: card.Heart, Rank: card.RankQ, Color: card.Red},
		{Suit: card.Joker, Rank: card.RankBlackJoker, Color: card.Black},
	}

	// Cards -> Infos -> Cards
	infos := CardsToInfos(originals)
	results := InfosToCards(infos)

	require.Len(t, results, len(originals))
	for i, orig := range originals {
		assert.Equal(t, orig, results[i], "Mismatch at index %d", i)
	}
}

func TestEmptyCards(t *testing.T) {
	t.Parallel()

	// Empty slice should work
	infos := CardsToInfos([]card.Card{})
	assert.Empty(t, infos)

	cards := InfosToCards([]protocol.CardInfo{})
	assert.Empty(t, cards)
}

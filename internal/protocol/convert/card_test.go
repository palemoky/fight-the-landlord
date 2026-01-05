package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/card"
)

func TestNewCardCounter(t *testing.T) {
	cc := NewCardCounter()

	require.NotNil(t, cc, "NewCardCounter should not return nil")

	remaining := cc.GetRemaining()

	// Test that all ranks from 3 to A have 4 cards
	for rank := card.Rank3; rank <= card.RankA; rank++ {
		assert.Equal(t, 4, remaining[rank], "Rank %v should have 4 cards", rank)
	}

	// Test Rank2 has 4 cards
	assert.Equal(t, 4, remaining[card.Rank2], "Rank2 should have 4 cards")

	// Test Jokers have 1 card each
	assert.Equal(t, 1, remaining[card.RankBlackJoker], "BlackJoker should have 1 card")
	assert.Equal(t, 1, remaining[card.RankRedJoker], "RedJoker should have 1 card")

	// Total should be 54 cards
	total := 0
	for _, count := range remaining {
		total += count
	}
	assert.Equal(t, 54, total, "Total cards should be 54")
}

func TestCardCounter_Reset(t *testing.T) {
	cc := NewCardCounter()

	// Deduct some cards
	testCards := []card.Card{
		{Rank: card.Rank3},
		{Rank: card.Rank3},
		{Rank: card.RankA},
	}
	cc.DeductCards(testCards)

	// Verify cards were deducted
	assert.Equal(t, 2, cc.GetRemaining()[card.Rank3], "After deducting 2 Rank3, should have 2 left")

	// Reset
	cc.Reset()

	// Verify all cards are back to original state
	remaining := cc.GetRemaining()
	assert.Equal(t, 4, remaining[card.Rank3], "After reset, Rank3 should have 4 cards")
	assert.Equal(t, 4, remaining[card.RankA], "After reset, RankA should have 4 cards")
}

func TestCardCounter_DeductCards(t *testing.T) {
	tests := []struct {
		name          string
		cardsToDeduct []card.Card
		expectedRank3 int
		expectedRankA int
	}{
		{
			name: "deduct single card",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
			},
			expectedRank3: 3,
			expectedRankA: 4,
		},
		{
			name: "deduct multiple same rank",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.Rank3},
				{Rank: card.Rank3},
			},
			expectedRank3: 1,
			expectedRankA: 4,
		},
		{
			name: "deduct all of one rank",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.Rank3},
				{Rank: card.Rank3},
				{Rank: card.Rank3},
			},
			expectedRank3: 0,
			expectedRankA: 4,
		},
		{
			name: "deduct mixed ranks",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.RankA},
				{Rank: card.Rank3},
			},
			expectedRank3: 2,
			expectedRankA: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cc := NewCardCounter()
			cc.DeductCards(tt.cardsToDeduct)

			remaining := cc.GetRemaining()
			assert.Equal(t, tt.expectedRank3, remaining[card.Rank3], "Rank3 count mismatch")
			assert.Equal(t, tt.expectedRankA, remaining[card.RankA], "RankA count mismatch")
		})
	}
}

func TestCardCounter_DeductCards_OverDeduct(t *testing.T) {
	cc := NewCardCounter()

	// Try to deduct more cards than available
	manyCards := make([]card.Card, 10)
	for i := range manyCards {
		manyCards[i] = card.Card{Rank: card.Rank3}
	}

	cc.DeductCards(manyCards)

	// Should not go below 0
	assert.GreaterOrEqual(t, cc.GetRemaining()[card.Rank3], 0, "Card count should not go below 0")
}

func TestCardCounter_DeductCards_EmptySlice(t *testing.T) {
	cc := NewCardCounter()
	initialRemaining := cc.GetRemaining()

	// Deduct empty slice
	cc.DeductCards([]card.Card{})

	// Nothing should change
	afterRemaining := cc.GetRemaining()
	for rank := card.Rank3; rank <= card.RankRedJoker; rank++ {
		assert.Equal(t, initialRemaining[rank], afterRemaining[rank],
			"Rank %v count should not change after deducting empty slice", rank)
	}
}

func TestCardCounter_GetRemaining(t *testing.T) {
	cc := NewCardCounter()
	remaining := cc.GetRemaining()

	require.NotNil(t, remaining, "GetRemaining should not return nil")

	// Verify it returns the actual map (not a copy)
	// Modifying the returned map should affect the counter
	originalCount := remaining[card.Rank3]
	remaining[card.Rank3] = 99

	assert.Equal(t, 99, cc.GetRemaining()[card.Rank3],
		"GetRemaining should return the actual map, not a copy")

	// Restore for other tests
	remaining[card.Rank3] = originalCount
}

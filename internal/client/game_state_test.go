package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

func TestNewGameState(t *testing.T) {
	gs := NewGameState()

	require.NotNil(t, gs, "NewGameState should not return nil")
	require.NotNil(t, gs.CardCounter, "CardCounter should be initialized")

	// Verify all fields are zero values
	assert.Nil(t, gs.Hand, "Hand should be nil")
	assert.Nil(t, gs.BottomCards, "BottomCards should be nil")
	assert.Nil(t, gs.Players, "Players should be nil")
	assert.Empty(t, gs.RoomCode, "RoomCode should be empty")
	assert.False(t, gs.IsLandlord, "IsLandlord should be false")
}

func TestGameState_SortHand(t *testing.T) {
	tests := []struct {
		name     string
		input    []card.Card
		expected []card.Rank
	}{
		{
			name: "already sorted",
			input: []card.Card{
				{Rank: card.RankA},
				{Rank: card.Rank5},
				{Rank: card.Rank3},
			},
			expected: []card.Rank{card.RankA, card.Rank5, card.Rank3},
		},
		{
			name: "unsorted ascending",
			input: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.Rank5},
				{Rank: card.RankA},
			},
			expected: []card.Rank{card.RankA, card.Rank5, card.Rank3},
		},
		{
			name: "with jokers",
			input: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.RankBlackJoker},
				{Rank: card.Rank2},
				{Rank: card.RankRedJoker},
			},
			expected: []card.Rank{card.RankRedJoker, card.RankBlackJoker, card.Rank2, card.Rank3},
		},
		{
			name: "duplicate ranks",
			input: []card.Card{
				{Rank: card.Rank5},
				{Rank: card.Rank3},
				{Rank: card.Rank5},
				{Rank: card.Rank3},
			},
			expected: []card.Rank{card.Rank5, card.Rank5, card.Rank3, card.Rank3},
		},
		{
			name:     "empty hand",
			input:    []card.Card{},
			expected: []card.Rank{},
		},
		{
			name: "single card",
			input: []card.Card{
				{Rank: card.Rank7},
			},
			expected: []card.Rank{card.Rank7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gs := NewGameState()
			gs.Hand = tt.input
			gs.SortHand()

			require.Len(t, gs.Hand, len(tt.expected), "Hand length mismatch")

			for i, expectedRank := range tt.expected {
				assert.Equal(t, expectedRank, gs.Hand[i].Rank,
					"Position %d: rank mismatch", i)
			}
		})
	}
}

func TestGameState_Reset(t *testing.T) {
	gs := NewGameState()

	// Set up some state
	gs.Hand = []card.Card{{Rank: card.Rank3}, {Rank: card.Rank5}}
	gs.BottomCards = []card.Card{{Rank: card.RankA}}
	gs.Players = []protocol.PlayerInfo{
		{ID: "player1", Name: "Alice"},
		{ID: "player2", Name: "Bob"},
	}
	gs.RoomCode = "ROOM123"
	gs.CurrentTurn = "player1"
	gs.LastPlayedBy = "player2"
	gs.LastPlayedName = "Bob"
	gs.LastPlayed = []card.Card{{Rank: card.Rank7}}
	gs.LastHandType = "Single"
	gs.IsLandlord = true
	gs.Winner = "player1"
	gs.WinnerIsLandlord = true

	// Deduct some cards from counter
	gs.CardCounter.DeductCards([]card.Card{{Rank: card.Rank3}})

	// Reset
	gs.Reset()

	// Verify all fields are cleared
	assert.Nil(t, gs.Hand, "Hand should be nil after reset")
	assert.Nil(t, gs.BottomCards, "BottomCards should be nil after reset")
	assert.Nil(t, gs.Players, "Players should be nil after reset")
	assert.Empty(t, gs.RoomCode, "RoomCode should be empty after reset")
	assert.Empty(t, gs.CurrentTurn, "CurrentTurn should be empty after reset")
	assert.Empty(t, gs.LastPlayedBy, "LastPlayedBy should be empty after reset")
	assert.Empty(t, gs.LastPlayedName, "LastPlayedName should be empty after reset")
	assert.Nil(t, gs.LastPlayed, "LastPlayed should be nil after reset")
	assert.Empty(t, gs.LastHandType, "LastHandType should be empty after reset")
	assert.False(t, gs.IsLandlord, "IsLandlord should be false after reset")
	assert.Empty(t, gs.Winner, "Winner should be empty after reset")
	assert.False(t, gs.WinnerIsLandlord, "WinnerIsLandlord should be false after reset")

	// Verify CardCounter is reset to full deck
	require.NotNil(t, gs.CardCounter, "CardCounter should not be nil after reset")
	remaining := gs.CardCounter.GetRemaining()
	assert.Equal(t, 4, remaining[card.Rank3],
		"CardCounter should be reset, Rank3 should have 4 cards")
}

func TestGameState_MultipleResets(t *testing.T) {
	gs := NewGameState()

	// Reset multiple times
	for i := 0; i < 5; i++ {
		gs.Hand = []card.Card{{Rank: card.Rank3}}
		gs.RoomCode = "TEST"
		gs.Reset()

		assert.Nil(t, gs.Hand, "Reset #%d: Hand should be nil", i+1)
		assert.Empty(t, gs.RoomCode, "Reset #%d: RoomCode should be empty", i+1)
	}
}

func TestGameState_SortHand_NilHand(t *testing.T) {
	gs := NewGameState()
	gs.Hand = nil

	// Should not panic
	assert.NotPanics(t, func() {
		gs.SortHand()
	}, "SortHand should not panic with nil hand")

	// Hand should still be nil
	assert.Nil(t, gs.Hand, "Hand should remain nil after sorting")
}

func TestGameState_Integration(t *testing.T) {
	// Test a realistic game flow
	gs := NewGameState()

	// Initial state
	require.NotNil(t, gs.CardCounter, "CardCounter not initialized")

	// Deal cards
	gs.Hand = []card.Card{
		{Rank: card.Rank3},
		{Rank: card.Rank7},
		{Rank: card.RankA},
		{Rank: card.Rank5},
	}

	// Sort hand
	gs.SortHand()
	assert.Equal(t, card.RankA, gs.Hand[0].Rank, "Hand not sorted correctly")

	// Update card counter
	gs.CardCounter.DeductCards(gs.Hand)
	assert.Equal(t, 3, gs.CardCounter.GetRemaining()[card.Rank3],
		"Expected 3 Rank3 remaining")

	// Set game info
	gs.RoomCode = "ABC123"
	gs.IsLandlord = true
	gs.Players = []protocol.PlayerInfo{
		{ID: "p1", Name: "Alice"},
		{ID: "p2", Name: "Bob"},
	}

	// Verify state
	assert.Equal(t, "ABC123", gs.RoomCode, "RoomCode not set correctly")
	assert.True(t, gs.IsLandlord, "IsLandlord not set correctly")
	assert.Len(t, gs.Players, 2, "Players not set correctly")

	// Reset for next game
	gs.Reset()
	assert.Empty(t, gs.RoomCode, "RoomCode should be empty after reset")
	assert.False(t, gs.IsLandlord, "IsLandlord should be false after reset")
	assert.Nil(t, gs.Players, "Players should be nil after reset")
	assert.Equal(t, 4, gs.CardCounter.GetRemaining()[card.Rank3],
		"CardCounter not reset properly")
}

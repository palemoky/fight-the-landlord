package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/rule"
)

// testCards is a helper function to quickly create cards for testing.
func testCards(ranks ...card.Rank) []card.Card {
	cards := make([]card.Card, len(ranks))
	for i, r := range ranks {
		cards[i] = card.Card{Rank: r, Suit: card.Spade, Color: card.Black}
	}
	return cards
}

// setupTestGame creates a predictable game state for testing, bypassing random elements.
func setupTestGame() *Game {
	g := NewGame()
	// Override players with predictable hands
	g.Players[0].Hand = testCards(card.Rank3, card.Rank4, card.Rank5, card.RankK, card.RankK)
	g.Players[1].Hand = testCards(card.Rank6, card.Rank7, card.Rank8, card.RankA, card.RankA)
	g.Players[2].Hand = testCards(card.Rank9, card.Rank10, card.RankJ, card.Rank2, card.Rank2)
	for _, p := range g.Players {
		p.SortHand()
	}
	g.CurrentTurn = 0
	g.LastPlayerIdx = 0
	return g
}

// TestPreprocessInput uses a table to test all timeout and input scenarios.
func TestPreprocessInput(t *testing.T) {
	testCases := []struct {
		name           string
		setupGame      func(g *Game) *Player // Setup game state and return the current player
		initialInput   string
		expectedOutput string
	}{
		{
			name: "with normal input string",
			setupGame: func(g *Game) *Player {
				return g.Players[0]
			},
			initialInput:   "KK",
			expectedOutput: "KK",
		},
		{
			name: "with timeout during free play",
			setupGame: func(g *Game) *Player {
				g.LastPlayedHand = rule.ParsedHand{} // New round
				player := g.Players[0]
				player.Hand = testCards(card.Rank5, card.Rank4, card.Rank3) // Smallest is 3
				player.SortHand()
				return player
			},
			initialInput:   "", // Timeout
			expectedOutput: "3",
		},
		{
			name: "with timeout when must beat a hand",
			setupGame: func(g *Game) *Player {
				g.LastPlayedHand, _ = rule.ParseHand(testCards(card.Rank6))
				g.CurrentTurn = 1
				g.LastPlayerIdx = 0
				return g.Players[1]
			},
			initialInput:   "", // Timeout
			expectedOutput: "PASS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := setupTestGame()
			player := tc.setupGame(g)

			processed := g.preprocessInput(tc.initialInput, player)
			assert.Equal(t, tc.expectedOutput, processed)
		})
	}
}

// TestHandlePass uses a table to test pass logic.
func TestHandlePass(t *testing.T) {
	testCases := []struct {
		name        string
		setupGame   func(g *Game)
		expectError bool
		assertState func(t *testing.T, g *Game) // Check game state after the call
	}{
		{
			name: "valid pass increments consecutive passes",
			setupGame: func(g *Game) {
				g.CurrentTurn = 1
				g.LastPlayerIdx = 0 // Player 0 played last
			},
			expectError: false,
			assertState: func(t *testing.T, g *Game) {
				assert.Equal(t, 1, g.ConsecutivePasses)
			},
		},
		{
			name: "invalid pass on a free play turn",
			setupGame: func(g *Game) {
				g.CurrentTurn = 0
				g.LastPlayerIdx = 0 // Player 0 is starting the round
			},
			expectError: true,
			assertState: func(t *testing.T, g *Game) {
				assert.Equal(t, 0, g.ConsecutivePasses) // State should not change
			},
		},
		{
			name: "two consecutive passes resets the round",
			setupGame: func(g *Game) {
				g.CurrentTurn = 2
				g.LastPlayerIdx = 0
				g.ConsecutivePasses = 1
				g.LastPlayedHand, _ = rule.ParseHand(testCards(card.RankK))
			},
			expectError: false,
			assertState: func(t *testing.T, g *Game) {
				assert.True(t, g.LastPlayedHand.IsEmpty())
				assert.Equal(t, (2+1)%3, g.LastPlayerIdx, "New round should start with the next player")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := setupTestGame()
			tc.setupGame(g)

			err := g.handlePass()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tc.assertState(t, g)
		})
	}
}

// TestHandlePlay uses a table to test play logic.
func TestHandlePlay(t *testing.T) {
	testCases := []struct {
		name        string
		setupGame   func(g *Game) *Player
		input       string
		expectError bool
		assertState func(t *testing.T, g *Game)
	}{
		{
			name: "valid play on a new round",
			setupGame: func(g *Game) *Player {
				player := g.Players[0]
				player.Hand = testCards(card.RankK, card.RankK)
				return player
			},
			input:       "KK",
			expectError: false,
			assertState: func(t *testing.T, g *Game) {
				assert.Len(t, g.Players[0].Hand, 0)
				assert.Equal(t, g.CurrentTurn, g.LastPlayerIdx)
				assert.Equal(t, rule.Pair, g.LastPlayedHand.Type)
			},
		},
		{
			name: "valid beat of a previous hand",
			setupGame: func(g *Game) *Player {
				g.LastPlayedHand, _ = rule.ParseHand(testCards(card.Rank10))
				g.LastPlayerIdx = 0
				player := g.Players[1]
				player.Hand = testCards(card.RankA)
				return player
			},
			input:       "A",
			expectError: false,
			assertState: func(t *testing.T, g *Game) {
				assert.Equal(t, card.RankA, g.LastPlayedHand.KeyRank)
			},
		},
		{
			name: "invalid beat (weaker hand)",
			setupGame: func(g *Game) *Player {
				g.LastPlayedHand, _ = rule.ParseHand(testCards(card.RankQ))
				g.LastPlayerIdx = 0
				g.CurrentTurn = 1
				player := g.Players[1]
				player.Hand = testCards(card.Rank6)
				return player
			},
			input:       "6",
			expectError: true,
			assertState: func(t *testing.T, g *Game) {
				assert.Equal(t, card.RankQ, g.LastPlayedHand.KeyRank)
				assert.Len(t, g.Players[1].Hand, 1)
			},
		},
		{
			name: "invalid play (cards not in hand)",
			setupGame: func(g *Game) *Player {
				return g.Players[0] // Hand is [3,4,5,K,K]
			},
			input:       "AA",
			expectError: true,
			assertState: func(t *testing.T, g *Game) {
				assert.True(t, g.LastPlayedHand.IsEmpty())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := setupTestGame()
			player := tc.setupGame(g)

			err := g.handlePlay(player, tc.input)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tc.assertState(t, g)
		})
	}
}

// TestAdvanceToNextTurn tests the logic for advancing the turn and setting the next player's state.
func TestAdvanceToNextTurn(t *testing.T) {
	testCases := []struct {
		name            string
		setupGame       func(g *Game)
		expectedTurn    int
		expectedCanPlay bool
	}{
		{
			name: "next turn is a free play",
			setupGame: func(g *Game) {
				g.CurrentTurn = 0
				g.LastPlayerIdx = 0
				g.LastPlayedHand = rule.ParsedHand{}
			},
			expectedTurn:    1,
			expectedCanPlay: true,
		},
		{
			name: "next player must beat and can",
			setupGame: func(g *Game) {
				g.CurrentTurn = 0
				g.LastPlayerIdx = 0
				g.LastPlayedHand, _ = rule.ParseHand(testCards(card.Rank3)) // Single 3
				// Player 1 has A which can beat 3
				g.Players[1].Hand = testCards(card.RankA)
			},
			expectedTurn:    1,
			expectedCanPlay: true,
		},
		{
			name: "next player must beat and cannot",
			setupGame: func(g *Game) {
				g.CurrentTurn = 1
				g.LastPlayerIdx = 1
				g.LastPlayedHand, _ = rule.ParseHand(testCards(card.RankA, card.RankA)) // Pair of A
				// Player 2 only has single cards, cannot beat a pair
				g.Players[2].Hand = testCards(card.Rank3, card.Rank4, card.Rank5)
			},
			expectedTurn:    2,
			expectedCanPlay: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := setupTestGame()
			tc.setupGame(g)

			g.advanceToNextTurn()

			assert.Equal(t, tc.expectedTurn, g.CurrentTurn)
			assert.Equal(t, tc.expectedCanPlay, g.CanCurrentPlayerPlay)
		})
	}
}

// TestPlayTurn_Integration provides a simple integration test for a full turn.
func TestPlayTurn_Integration(t *testing.T) {
	// A more linear test is better here than a complex table.
	g := setupTestGame()

	// Override hands for a predictable game flow
	g.Players[0].Hand = testCards(card.Rank3, card.RankK)
	g.Players[1].Hand = testCards(card.Rank4)
	g.Players[2].Hand = testCards(card.Rank5)
	g.CurrentTurn = 0

	// Player 0 plays 3
	err := g.PlayTurn("3")
	require.NoError(t, err)
	assert.Equal(t, 1, g.CurrentTurn, "Turn should advance to Player 1")
	assert.Equal(t, card.Rank3, g.LastPlayedHand.KeyRank)
	assert.Len(t, g.Players[0].Hand, 1)

	// Player 1 plays 4 (beating 3)
	err = g.PlayTurn("4")
	require.NoError(t, err)
	assert.Equal(t, 2, g.CurrentTurn, "Turn should advance to Player 2")
	assert.Equal(t, card.Rank4, g.LastPlayedHand.KeyRank)
	assert.Len(t, g.Players[1].Hand, 0) // Player 1 wins

	// Check for winner
	winner, isOver := g.CheckWinner()
	require.True(t, isOver)
	assert.Equal(t, g.Players[1].Name, winner.Name)
}

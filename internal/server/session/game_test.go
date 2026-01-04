package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/apperrors"
	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/game/room"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

func TestHandleBid_Success(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// Get current bidder
	currentBidder := gs.players[gs.currentBidder]

	// Bid successfully
	err := gs.HandleBid(currentBidder.ID, true)
	require.NoError(t, err)

	// Verify landlord is set
	assert.Equal(t, GameStatePlaying, gs.state)
	assert.True(t, currentBidder.IsLandlord)
}

func TestHandleBid_NotYourTurn(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// Try to bid with wrong player
	wrongPlayer := gs.players[(gs.currentBidder+1)%3]
	err := gs.HandleBid(wrongPlayer.ID, true)
	assert.ErrorIs(t, err, apperrors.ErrNotYourTurn)
}

func TestHandleBid_GameNotStarted(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	// Don't start the game

	err := gs.HandleBid("p1", true)
	assert.ErrorIs(t, err, apperrors.ErrGameNotStart)
}

func TestHandleBid_AllPass_RandomLandlord(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// All players pass
	for range 3 {
		currentBidder := gs.players[gs.currentBidder]
		err := gs.HandleBid(currentBidder.ID, false)
		require.NoError(t, err)
	}

	// Verify a landlord was randomly assigned
	assert.Equal(t, GameStatePlaying, gs.state)
	landlordCount := 0
	for _, p := range gs.players {
		if p.IsLandlord {
			landlordCount++
		}
	}
	assert.Equal(t, 1, landlordCount)
}

func TestHandlePlayCards_Success(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// Set landlord and start playing
	gs.mu.Lock()
	gs.state = GameStatePlaying
	gs.currentPlayer = 0
	gs.players[0].IsLandlord = true
	// Give player some cards
	gs.players[0].Hand = []card.Card{
		{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
		{Suit: card.Heart, Rank: card.Rank3, Color: card.Red},
		{Suit: card.Diamond, Rank: card.Rank3, Color: card.Red},
	}
	gs.mu.Unlock()

	// Play cards
	cardsToPlay := []protocol.CardInfo{
		convert.CardToInfo(gs.players[0].Hand[0]),
		convert.CardToInfo(gs.players[0].Hand[1]),
		convert.CardToInfo(gs.players[0].Hand[2]),
	}

	err := gs.HandlePlayCards("p1", cardsToPlay)
	require.NoError(t, err)

	// Verify cards were removed
	assert.Len(t, gs.players[0].Hand, 0)
}

func TestHandlePlayCards_NotYourTurn(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	gs.mu.Lock()
	gs.state = GameStatePlaying
	gs.currentPlayer = 0
	gs.mu.Unlock()

	// Try to play with wrong player
	err := gs.HandlePlayCards("p2", []protocol.CardInfo{})
	assert.ErrorIs(t, err, apperrors.ErrNotYourTurn)
}

func TestHandlePlayCards_InvalidCards(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	gs.mu.Lock()
	gs.state = GameStatePlaying
	gs.currentPlayer = 0
	gs.players[0].Hand = []card.Card{
		{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
	}
	gs.mu.Unlock()

	// Try to play cards not in hand
	invalidCards := []protocol.CardInfo{
		{Suit: int(card.Heart), Rank: int(card.RankA), Color: int(card.Red)},
	}

	err := gs.HandlePlayCards("p1", invalidCards)
	assert.ErrorIs(t, err, apperrors.ErrInvalidCards)
}

func TestHandlePass_Success(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	gs.mu.Lock()
	gs.state = GameStatePlaying
	gs.currentPlayer = 1
	gs.lastPlayerIdx = 0 // Player 0 played last
	gs.mu.Unlock()

	// Pass successfully
	err := gs.HandlePass("p2")
	require.NoError(t, err)

	// Verify turn moved to next player
	assert.Equal(t, 2, gs.currentPlayer)
}

func TestHandlePass_MustPlay(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	gs.mu.Lock()
	gs.state = GameStatePlaying
	gs.currentPlayer = 0
	gs.lastPlayerIdx = 0 // Same player - must play
	gs.mu.Unlock()

	// Try to pass when must play
	err := gs.HandlePass("p1")
	assert.ErrorIs(t, err, apperrors.ErrMustPlay)
}

func TestHandlePass_TwoPassesNewRound(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	gs.mu.Lock()
	gs.state = GameStatePlaying
	gs.currentPlayer = 1
	gs.lastPlayerIdx = 0
	gs.consecutivePasses = 1 // Already one pass
	gs.mu.Unlock()

	// Second pass should trigger new round
	err := gs.HandlePass("p2")
	require.NoError(t, err)

	// Verify new round started
	assert.Equal(t, 0, gs.consecutivePasses)
	assert.True(t, gs.lastPlayedHand.IsEmpty())
}

func TestValidateCardsInHand(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))

	player := &GamePlayer{
		Hand: []card.Card{
			{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
			{Suit: card.Heart, Rank: card.Rank4, Color: card.Red},
			{Suit: card.Diamond, Rank: card.Rank5, Color: card.Red},
		},
	}

	tests := []struct {
		name  string
		cards []card.Card
		valid bool
	}{
		{
			name: "Valid cards",
			cards: []card.Card{
				{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
			},
			valid: true,
		},
		{
			name: "Multiple valid cards",
			cards: []card.Card{
				{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
				{Suit: card.Heart, Rank: card.Rank4, Color: card.Red},
			},
			valid: true,
		},
		{
			name: "Invalid card not in hand",
			cards: []card.Card{
				{Suit: card.Club, Rank: card.RankA, Color: card.Black},
			},
			valid: false,
		},
		{
			name: "Duplicate card (only one in hand)",
			cards: []card.Card{
				{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
				{Suit: card.Spade, Rank: card.Rank3, Color: card.Black},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := gs.validateCardsInHand(player, tt.cards)
			assert.Equal(t, tt.valid, result)
		})
	}
}

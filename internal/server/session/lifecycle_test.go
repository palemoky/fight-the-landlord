package session

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/game/room"
	"github.com/palemoky/fight-the-landlord/internal/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

func TestStartGame_DealCards(t *testing.T) {
	t.Parallel()

	// Setup room with 3 players
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	// Create game session
	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))

	// Start game
	gs.Start()

	// Verify state
	assert.Equal(t, GameStateBidding, gs.state)
	assert.Equal(t, room.RoomStateBidding, r.State)

	// Verify each player has 17 cards
	for i, p := range gs.players {
		assert.Len(t, p.Hand, 17, "Player %d should have 17 cards", i)
	}

	// Verify bottom cards (3 cards)
	assert.Len(t, gs.bottomCards, 3)

	// Verify total cards = 54
	totalCards := len(gs.bottomCards)
	for _, p := range gs.players {
		totalCards += len(p.Hand)
	}
	assert.Equal(t, 54, totalCards)
}

func TestStartGame_CardsAreSorted(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// Verify hands are sorted (descending by rank)
	for _, p := range gs.players {
		for i := 0; i < len(p.Hand)-1; i++ {
			assert.GreaterOrEqual(t, p.Hand[i].Rank, p.Hand[i+1].Rank,
				"Cards should be sorted in descending order")
		}
	}
}

func TestStartGame_BidderSelected(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// Verify a bidder was selected (0, 1, or 2)
	assert.GreaterOrEqual(t, gs.currentBidder, 0)
	assert.Less(t, gs.currentBidder, 3)
}

func TestEndGame_WinnerAnnounced(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))
	gs.Start()

	// Set winner
	winner := gs.players[0]
	winner.IsLandlord = true

	// End game
	gs.endGame(winner)

	// Verify state
	assert.Equal(t, GameStateEnded, gs.state)
	assert.Equal(t, room.RoomStateEnded, r.State)
}

func TestNewGameSession_Initialization(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	// Create session
	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))

	// Verify initialization
	assert.Equal(t, GameStateInit, gs.state)
	assert.Len(t, gs.players, 3)
	assert.Equal(t, -1, gs.highestBidder)

	// Verify players are in correct seats
	for i, p := range gs.players {
		assert.Equal(t, i, p.Seat)
		assert.Equal(t, r.PlayerOrder[i], p.ID)
	}
}

func TestGameSession_PlayerOfflineHandling(t *testing.T) {
	t.Parallel()

	// Setup
	r := room.NewMockRoom("TEST123", testutil.NewSimpleClient("p1", "Player1"))
	r.Players["p2"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p2", "Player2"), Seat: 1}
	r.Players["p3"] = &room.RoomPlayer{Client: testutil.NewSimpleClient("p3", "Player3"), Seat: 2}
	r.PlayerOrder = []string{"p1", "p2", "p3"}

	gs := NewGameSession(r, storage.NewLeaderboardManager(nil))

	// Mark player as offline
	gs.mu.Lock()
	gs.players[0].IsOffline = true
	gs.mu.Unlock()

	// Verify offline status
	gs.mu.RLock()
	assert.True(t, gs.players[0].IsOffline)
	gs.mu.RUnlock()
}

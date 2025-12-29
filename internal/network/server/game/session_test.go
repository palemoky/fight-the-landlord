package game

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// Helper to create a room with 3 mock clients
func newTestRoom() *Room {
	room := &Room{
		Code:        "123",
		Players:     make(map[string]*RoomPlayer),
		PlayerOrder: []string{"p1", "p2", "p3"},
		State:       RoomStateWaiting,
	}

	clients := []*MockClient{
		{ID: "p1", Name: "Player1", RoomCode: "123"},
		{ID: "p2", Name: "Player2", RoomCode: "123"},
		{ID: "p3", Name: "Player3", RoomCode: "123"},
	}

	for i, c := range clients {
		room.Players[c.ID] = &RoomPlayer{
			Client: c, // Implements ClientInterface
			Seat:   i,
			Ready:  true,
		}
	}

	return room
}

// MockLeaderboard for GameSession recordGameResults
type mockLeaderboard struct {
	mock.Mock
}

func (m *mockLeaderboard) RecordGameResult(ctx context.Context, playerID, playerName string, isLandlord, isWinner bool) error {
	args := m.Called(ctx, playerID, playerName, isLandlord, isWinner)
	return args.Error(0)
}

func (m *mockLeaderboard) GetPlayerStats(ctx context.Context, playerID string) (interface{}, error) {
	return nil, nil
}
func (m *mockLeaderboard) SavePlayerStats(ctx context.Context, stats interface{}) error   { return nil }
func (m *mockLeaderboard) UpdateLeaderboard(ctx context.Context, stats interface{}) error { return nil }
func (m *mockLeaderboard) GetLeaderboard(ctx context.Context, limit int) ([]interface{}, error) {
	return nil, nil
}

func (m *mockLeaderboard) GetPlayerRank(ctx context.Context, playerID string) (int64, error) {
	return 0, nil
}

func TestGameSession_Start(t *testing.T) {
	room := newTestRoom()
	gs := NewGameSession(room)
	room.game = gs

	gs.Start()

	assert.Equal(t, GameStateBidding, gs.state)
	assert.Len(t, gs.players, 3)

	// Check card distribution
	for _, p := range gs.players {
		assert.Len(t, p.Hand, 17)
	}
	assert.Len(t, gs.bottomCards, 3)

	// Check notification (MsgDealCards, MsgBidTurn)
	// Just check if messsages were sent to p1
	p1Client := room.Players["p1"].Client.(*MockClient)
	assert.NotEmpty(t, p1Client.Messages)

	hasDeal := false
	hasBid := false
	for _, msg := range p1Client.Messages {
		if msg.Type == protocol.MsgDealCards {
			hasDeal = true
		}
		if msg.Type == protocol.MsgBidTurn {
			hasBid = true
		}
	}
	assert.True(t, hasDeal)
	assert.True(t, hasBid)
}

func TestGameSession_HandleBid(t *testing.T) {
	room := newTestRoom()
	gs := NewGameSession(room)
	room.game = gs
	gs.Start()

	// Find current bidder
	bidderIdx := gs.currentBidder
	bidderID := gs.players[bidderIdx].ID

	// Valid Bid
	err := gs.HandleBid(bidderID, true)
	assert.NoError(t, err)

	// Since they bid, they become landlord and game starts
	assert.Equal(t, GameStatePlaying, gs.state)
	assert.Equal(t, bidderIdx, gs.highestBidder)
	assert.True(t, gs.players[bidderIdx].IsLandlord)
	assert.Len(t, gs.players[bidderIdx].Hand, 20) // 17 + 3
}

func TestGameSession_HandleBid_AllPass(t *testing.T) {
	room := newTestRoom()
	gs := NewGameSession(room)
	room.game = gs
	gs.Start()

	// Force fixed start
	gs.currentBidder = 0

	// 1. Pass
	err := gs.HandleBid("p1", false)
	assert.NoError(t, err)
	assert.Equal(t, GameStateBidding, gs.state)
	assert.Equal(t, 1, gs.currentBidder)

	// 2. Pass
	err = gs.HandleBid("p2", false)
	assert.NoError(t, err)
	assert.Equal(t, GameStateBidding, gs.state)
	assert.Equal(t, 2, gs.currentBidder)

	// 3. Pass -> Force Landlord (randomly picked among 3)
	err = gs.HandleBid("p3", false)
	assert.NoError(t, err)

	// Should be Playing now
	assert.Equal(t, GameStatePlaying, gs.state)
	assert.NotEqual(t, -1, gs.highestBidder)
}

func TestGameSession_HandlePlayCards(t *testing.T) {
	room := newTestRoom()
	gs := NewGameSession(room)
	room.game = gs
	gs.Start()

	// Make p1 landlord for predictability
	gs.currentBidder = 0
	gs.HandleBid("p1", true)

	assert.Equal(t, GameStatePlaying, gs.state)
	assert.Equal(t, 0, gs.currentPlayer) // Landlord starts

	p1 := gs.players[0]
	// Pick a valid card from hand to play
	cardToPlay := p1.Hand[0]
	// Play Valid Card
	playCards := []protocol.CardInfo{
		{Suit: int(cardToPlay.Suit), Rank: int(cardToPlay.Rank), Color: int(cardToPlay.Color)},
	}

	// Play Single Card
	err := gs.HandlePlayCards("p1", playCards)
	assert.NoError(t, err)

	// Verify state update
	assert.Equal(t, 19, len(p1.Hand))
	assert.Equal(t, 1, gs.currentPlayer) // Next turn
}

// Satisfy ServerContext interface fully for the mock
type MockServerContext struct {
	mock.Mock
}

func (m *MockServerContext) GetRedisStore() types.RedisStoreInterface { return nil }
func (m *MockServerContext) GetLeaderboard() types.LeaderboardInterface {
	args := m.Called()
	return args.Get(0).(types.LeaderboardInterface)
}
func (m *MockServerContext) GetSessionManager() types.SessionManagerInterface { return nil }
func (m *MockServerContext) GetRoomManager() types.RoomManagerInterface {
	args := m.Called()
	return args.Get(0).(types.RoomManagerInterface)
}
func (m *MockServerContext) GetMatcher() types.MatcherInterface                     { return nil }
func (m *MockServerContext) GetChatLimiter() types.ChatLimiterInterface             { return nil }
func (m *MockServerContext) GetClientByID(id string) types.ClientInterface          { return nil }
func (m *MockServerContext) RegisterClient(id string, client types.ClientInterface) {}
func (m *MockServerContext) UnregisterClient(id string)                             {}
func (m *MockServerContext) IsMaintenanceMode() bool                                { return false }
func (m *MockServerContext) GetOnlineCount() int                                    { return 0 }
func (m *MockServerContext) Broadcast(msg *protocol.Message)                        {}
func (m *MockServerContext) BroadcastToLobby(msg *protocol.Message)                 {}

// Satisfy RoomManagerInterface for LeaveRoom called in endGame

func TestGameSession_WinCondition(t *testing.T) {
	room := newTestRoom()

	// Setup Mock Server
	mockServer := new(MockServerContext)
	mockLdb := new(mockLeaderboard)

	room.server = mockServer // Accessing private field in same package

	gs := NewGameSession(room)
	room.game = gs
	gs.Start()

	// p1 becomes landlord
	gs.currentBidder = 0
	gs.HandleBid("p1", true)

	// Mock Expectations
	mockServer.On("GetLeaderboard").Return(mockLdb)

	// Expect record game result 3 times (p1 win, p2 loss, p3 loss)
	mockLdb.On("RecordGameResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3)

	// Cheat: Empty p1's hand except 1 card
	p1 := gs.players[0]
	lastCard := p1.Hand[0]
	p1.Hand = []card.Card{lastCard}

	// Play last card
	playCards := []protocol.CardInfo{{Suit: int(lastCard.Suit), Rank: int(lastCard.Rank), Color: int(lastCard.Color)}}
	err := gs.HandlePlayCards("p1", playCards)
	assert.NoError(t, err)

	assert.Equal(t, GameStateEnded, gs.state)

	// Verify interactions
	mockServer.AssertExpectations(t)
	mockLdb.AssertExpectations(t)
}

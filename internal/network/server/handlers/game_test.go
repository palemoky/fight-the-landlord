package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
)

// Helper to create a room with a running game session and mock clients
func setupGameRoom(t *testing.T) (*game.Room, []*MockClient) {
	room := &game.Room{
		Code:        "123",
		Players:     make(map[string]*game.RoomPlayer),
		PlayerOrder: []string{"p1", "p2", "p3"},
	}

	clients := make([]*MockClient, 3)
	for i := 0; i < 3; i++ {
		c := new(MockClient)
		id := room.PlayerOrder[i]
		c.On("GetID").Return(id)
		c.On("GetName").Return("Player" + id)
		c.On("GetRoom").Return("123")
		// Unexpected calls allowed for setup
		c.On("SetRoom", mock.Anything).Maybe()
		c.On("Close").Maybe()
		c.On("SendMessage", mock.Anything).Maybe()

		room.Players[id] = &game.RoomPlayer{
			Client: c,
			Seat:   i,
			Ready:  true,
		}
		clients[i] = c
	}

	// Create and start session
	gs := game.NewGameSession(room)
	room.SetGameSession(gs)
	gs.Start()

	return room, clients
}

func TestHandler_HandleBid_Success(t *testing.T) {
	room, clients := setupGameRoom(t)
	gs := room.GetGameSession() // Should be non-nil assuming we set it
	assert.NotNil(t, gs)

	mockServer := new(MockServer)
	mockRM := new(MockRoomManager)
	mockServer.On("GetRoomManager").Return(mockRM)
	mockRM.On("GetRoom", "123").Return(room)

	h := NewHandler(mockServer)

	success := false
	for _, c := range clients {
		payload := protocol.BidPayload{Bid: true}
		payloadBytes, _ := convert.EncodePayload(protocol.MsgBid, payload)
		msg := &protocol.Message{
			Type:    protocol.MsgBid,
			Payload: payloadBytes,
		}

		// Capture call count before
		callsBefore := len(c.Calls)

		h.handleBid(c, msg)

		// If success, gs state changes to Playing
		if gs.GetStateForSerialization() == game.GameStatePlaying {
			success = true
			break
		} else if len(c.Calls) > callsBefore {
			// Check if we got an error message
			lastCall := c.Calls[len(c.Calls)-1]
			if lastCall.Method == "SendMessage" {
				msgSent := lastCall.Arguments.Get(0).(*protocol.Message)
				if msgSent.Type == protocol.MsgError {
					var errP protocol.ErrorPayload
					if err := convert.DecodePayload(protocol.MsgError, msgSent.Payload, &errP); err != nil {
						t.Logf("Failed to decode error payload: %v", err)
					} else {
						t.Logf("Client %s bid failed: Code=%d Msg=%s", c.GetID(), errP.Code, errP.Message)
					}
				}
			}
		}
	}
	assert.True(t, success, "One player should successfully bid")
}

func TestHandler_HandlePlayCards_Success(t *testing.T) {
	room, clients := setupGameRoom(t)
	gs := room.GetGameSession()
	assert.NotNil(t, gs)

	mockServer := new(MockServer)
	mockRM := new(MockRoomManager)
	mockServer.On("GetRoomManager").Return(mockRM)
	mockRM.On("GetRoom", "123").Return(room)

	h := NewHandler(mockServer)

	// Force bidding phase to pass by simulating valid bids
	mockLdb := new(MockLeaderboard)
	mockServer.On("GetLeaderboard").Return(mockLdb)
	// Wait, mockLdb needed for recordGameResults if game ends? PlayCards might end game.
	// The previous test WinCondition in session_test.go had it.
	// Here we just play 1 card, not winning yet (starts with 20 cards).

	// Find bidder and bid
	currentTurnID := ""
	for _, c := range clients {
		payload := protocol.BidPayload{Bid: true}
		payloadBytes, _ := convert.EncodePayload(protocol.MsgBid, payload)
		h.handleBid(c, &protocol.Message{Type: protocol.MsgBid, Payload: payloadBytes})

		if gs.GetStateForSerialization() == game.GameStatePlaying {
			currentTurnID = c.GetID()
			break
		}
	}
	assert.NotEmpty(t, currentTurnID, "Should find a bidder")

	// Identify Landlord Client
	landlordIdx := gs.GetCurrentPlayerForSerialization()
	landlordID := room.PlayerOrder[landlordIdx]

	var landlordClient *MockClient
	for _, c := range clients {
		if c.GetID() == landlordID {
			landlordClient = c
			break
		}
	}
	assert.NotNil(t, landlordClient)

	// Play cards
	players := gs.GetPlayersForSerialization()
	landlordPlayer := players[landlordIdx]

	cardToPlay := landlordPlayer.Hand[0]
	playPayload := protocol.PlayCardsPayload{
		Cards: []protocol.CardInfo{
			{Suit: int(cardToPlay.Suit), Rank: int(cardToPlay.Rank), Color: int(cardToPlay.Color)},
		},
	}
	payloadBytes, _ := convert.EncodePayload(protocol.MsgPlayCards, playPayload)
	msg := &protocol.Message{Type: protocol.MsgPlayCards, Payload: payloadBytes}

	h.handlePlayCards(landlordClient, msg)

	// Verify hand size decreased
	assert.Equal(t, 19, len(landlordPlayer.Hand))
}

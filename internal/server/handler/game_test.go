package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	r "github.com/palemoky/fight-the-landlord/internal/game/room"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
	"github.com/palemoky/fight-the-landlord/internal/server/session"
	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

// Helper to create a room with a running game session and mock clients
func setupGameRoom(t *testing.T) (*r.Room, *session.GameSession, []*testutil.MockClient) {
	t.Helper()
	room := &r.Room{
		Code:        "123",
		Players:     make(map[string]*r.RoomPlayer),
		PlayerOrder: []string{"p1", "p2", "p3"},
	}

	clients := make([]*testutil.MockClient, 3)
	for i := range 3 {
		c := new(testutil.MockClient)
		id := room.PlayerOrder[i]
		c.On("GetID").Return(id)
		c.On("GetName").Return("Player" + id)
		c.On("GetRoom").Return("123")
		// Unexpected calls allowed for setup
		c.On("SetRoom", mock.Anything).Maybe()
		c.On("Close").Maybe()
		c.On("SendMessage", mock.Anything).Maybe()

		room.Players[id] = &r.RoomPlayer{
			Client: c,
			Seat:   i,
			Ready:  true,
		}
		clients[i] = c
	}

	// Create and start session
	gs := session.NewGameSession(room, nil)
	gs.Start()

	return room, gs, clients
}

func TestHandler_HandleBid_Success(t *testing.T) {
	room, gs, clients := setupGameRoom(t)

	mockServer := new(testutil.MockServer)
	mockRM := new(r.MockRoomManager)
	mockServer.On("GetRoomManager").Return(mockRM)
	mockRM.On("GetRoom", "123").Return(room)

	h := NewHandler(HandlerDeps{Server: mockServer})
	h.SetGameSession(room.Code, gs)

	assert.NotNil(t, h.GetGameSession(room.Code))

	success := false
	for _, c := range clients {
		payload := protocol.BidPayload{Bid: true}
		payloadBytes, _ := payloadconv.EncodePayload(protocol.MsgBid, payload)
		msg := &protocol.Message{
			Type:    protocol.MsgBid,
			Payload: payloadBytes,
		}

		// Capture call count before
		callsBefore := len(c.Calls)

		h.handleBid(c, msg)

		// 成功情况：状态改变
		if gs.GetStateForSerialization() == session.GameStatePlaying {
			success = true
			break
		}

		// 无响应情况：直接跳过
		if len(c.Calls) <= callsBefore {
			continue
		}

		// 失败情况：尝试解析并记录错误
		lastCall := c.Calls[len(c.Calls)-1]
		if lastCall.Method == "SendMessage" {
			if msgSent, ok := lastCall.Arguments.Get(0).(*protocol.Message); ok && msgSent.Type == protocol.MsgError {
				logErrorPayload(t, c.GetID(), msgSent)
			}
		}
	}
	assert.True(t, success, "One player should successfully bid")
}

func TestHandler_HandlePlayCards_Success(t *testing.T) {
	room, gs, clients := setupGameRoom(t)

	mockServer := new(testutil.MockServer)
	mockRM := new(r.MockRoomManager)
	mockServer.On("GetRoomManager").Return(mockRM)
	mockRM.On("GetRoom", "123").Return(room)

	h := NewHandler(HandlerDeps{Server: mockServer})
	h.SetGameSession(room.Code, gs)

	assert.NotNil(t, h.GetGameSession(room.Code))

	// Force bidding phase to pass by simulating valid bids
	mockLdb := new(testutil.MockLeaderboard)
	mockServer.On("GetLeaderboard").Return(mockLdb)
	// Wait, mockLdb needed for recordGameResults if game ends? PlayCards might end game.
	// The previous test WinCondition in session_test.go had it.
	// Here we just play 1 card, not winning yet (starts with 20 cards).

	// Find bidder and bid
	currentTurnID := ""
	for _, c := range clients {
		payload := protocol.BidPayload{Bid: true}
		payloadBytes, _ := payloadconv.EncodePayload(protocol.MsgBid, payload)
		h.handleBid(c, &protocol.Message{Type: protocol.MsgBid, Payload: payloadBytes})

		if gs.GetStateForSerialization() == session.GameStatePlaying {
			currentTurnID = c.GetID()
			break
		}
	}
	assert.NotEmpty(t, currentTurnID, "Should find a bidder")

	// Identify Landlord Client
	landlordIdx := gs.GetCurrentPlayerForSerialization()
	landlordID := room.PlayerOrder[landlordIdx]

	var landlordClient *testutil.MockClient
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
	payloadBytes, _ := payloadconv.EncodePayload(protocol.MsgPlayCards, playPayload)
	msg := &protocol.Message{Type: protocol.MsgPlayCards, Payload: payloadBytes}

	h.handlePlayCards(landlordClient, msg)

	// Verify hand size decreased
	assert.Equal(t, 19, len(landlordPlayer.Hand))
}

func logErrorPayload(t *testing.T, clientID string, msg *protocol.Message) {
	t.Helper()
	var errP protocol.ErrorPayload
	if err := payloadconv.DecodePayload(protocol.MsgError, msg.Payload, &errP); err != nil {
		t.Logf("Failed to decode error payload: %v", err)
	} else {
		t.Logf("Client %s bid failed: Code=%d Msg=%s", clientID, errP.Code, errP.Message)
	}
}

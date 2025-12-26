package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// Helper to create a fake Message
func createMessage(msgType protocol.MessageType, payload interface{}) *protocol.Message {
	data, _ := protocol.EncodePayload(msgType, payload)
	return &protocol.Message{
		Type:    msgType,
		Payload: data,
	}
}

func TestHandleMsgRoomCreated(t *testing.T) {
	// Setup
	model := NewOnlineModel("ws://localhost:8080")
	player := protocol.PlayerInfo{ID: "p1", Name: "Player 1"}
	payload := protocol.RoomCreatedPayload{
		RoomCode: "1234",
		Player:   player,
	}
	msg := createMessage(protocol.MsgRoomCreated, payload)

	// Execute
	model.handleServerMessage(msg)

	// Verify
	assert.Equal(t, "1234", model.game.roomCode)
	assert.Equal(t, PhaseWaiting, model.phase)
	assert.Len(t, model.game.players, 1)
	assert.Equal(t, "p1", model.game.players[0].ID)
	assert.Equal(t, "输入 R 准备", model.input.Placeholder)
}

func TestHandleMsgGameStart(t *testing.T) {
	// Setup
	model := NewOnlineModel("ws://localhost:8080")
	players := []protocol.PlayerInfo{
		{ID: "p1", Name: "Player 1"},
		{ID: "p2", Name: "Player 2"},
		{ID: "p3", Name: "Player 3"},
	}
	payload := protocol.GameStartPayload{
		Players: players,
	}
	msg := createMessage(protocol.MsgGameStart, payload)

	// Execute
	model.handleServerMessage(msg)

	// Verify
	assert.Len(t, model.game.players, 3)
	assert.Equal(t, "p2", model.game.players[1].ID)
}

func TestHandleMsgDealCards(t *testing.T) {
	// Setup
	model := NewOnlineModel("ws://localhost:8080")
	cards := []protocol.CardInfo{
		{Rank: int(card.Rank3), Suit: int(card.Spade)},
		{Rank: int(card.RankA), Suit: int(card.Heart)},
		{Rank: int(card.Rank2), Suit: int(card.Club)},
	}
	payload := protocol.DealCardsPayload{
		Cards: cards,
	}
	msg := createMessage(protocol.MsgDealCards, payload)

	// Pre-requisite: Setup players for remaining card calculation logic
	model.game.players = []protocol.PlayerInfo{{ID: "p1"}}

	// Execute
	model.handleServerMessage(msg)

	// Verify
	assert.Len(t, model.game.hand, 3)
	// Check sorting (2 > A > 3)
	assert.Equal(t, card.Rank2, model.game.hand[0].Rank)
	assert.Equal(t, card.RankA, model.game.hand[1].Rank)
	assert.Equal(t, card.Rank3, model.game.hand[2].Rank)

	// Check remaining cards initialization
	assert.NotNil(t, model.game.remainingCards)
	// Example: 2s should be 4 (total) - 1 (in hand) = 3
	assert.Equal(t, 3, model.game.remainingCards[card.Rank2])
}

func TestHandleMsgPlayTurn(t *testing.T) {
	// Setup
	model := NewOnlineModel("ws://localhost:8080")
	model.playerID = "p1"
	model.game.players = []protocol.PlayerInfo{
		{ID: "p1", Name: "User"},
		{ID: "p2", Name: "Other"},
	}

	// Case 1: My Turn, Must Play
	payload := protocol.PlayTurnPayload{
		PlayerID: "p1",
		MustPlay: true,
		Timeout:  15,
	}
	msg := createMessage(protocol.MsgPlayTurn, payload)

	model.handleServerMessage(msg)

	assert.Equal(t, PhasePlaying, model.phase)
	assert.Equal(t, "p1", model.game.currentTurn)
	assert.Equal(t, "你必须出牌 (如 33344)", model.input.Placeholder)
	assert.True(t, model.input.Focused())
	assert.Equal(t, 15*time.Second, model.game.timerDuration)

	// Case 2: Other's Turn
	payloadOther := protocol.PlayTurnPayload{
		PlayerID: "p2",
		Timeout:  15,
	}
	msgOther := createMessage(protocol.MsgPlayTurn, payloadOther)

	model.handleServerMessage(msgOther)

	assert.Equal(t, "p2", model.game.currentTurn)
	// Placeholder logic iterates players to find name
	assert.Contains(t, model.input.Placeholder, "等待 Other 出牌")
	assert.False(t, model.input.Focused())
}

package handlers

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
)

func TestHandler_HandleChat_Lobby(t *testing.T) {
	// 1. Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLimiter := new(MockChatLimiter)

	h := NewHandler(mockServer)

	// 2. Expectations
	mockServer.On("GetChatLimiter").Return(mockLimiter)

	// For Lobby chat:
	mockClient.On("GetID").Return("p1")
	mockClient.On("GetName").Return("Player1")
	mockLimiter.On("AllowChat", "p1").Return(true, "")

	// Expect BroadcastToLobby to be called with a MsgChat message
	mockServer.On("BroadcastToLobby", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgChat
	})).Return()

	// 3. Execution
	payload := protocol.ChatPayload{
		Content: "Hello World",
		Scope:   "lobby",
	}
	msg := codec.MustNewMessage(protocol.MsgChat, payload)

	h.handleChat(mockClient, msg)

	// 4. Verification
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockLimiter.AssertExpectations(t)
}

func TestHandler_HandleChat_RateLimited(t *testing.T) {
	// 1. Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLimiter := new(MockChatLimiter)

	h := NewHandler(mockServer)

	// 2. Expectations
	mockServer.On("GetChatLimiter").Return(mockLimiter)
	mockClient.On("GetID").Return("p1")

	// Reject chat
	mockLimiter.On("AllowChat", "p1").Return(false, "Too fast")

	// Expect error message sent to client
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgError
	})).Return()

	// 3. Execution
	payload := protocol.ChatPayload{Content: "Spam"}
	msg := codec.MustNewMessage(protocol.MsgChat, payload)

	h.handleChat(mockClient, msg)

	// 4. Verification
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockLimiter.AssertExpectations(t)
}

func TestHandler_HandleChat_Room(t *testing.T) {
	// 1. Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLimiter := new(MockChatLimiter)
	mockRM := new(MockRoomManager)

	h := NewHandler(mockServer)

	// Use NewMockRoom helper which returns a real *game.Room
	room := NewMockRoom("123", nil)
	// Add p1 to room explicitly or use helper? behavior
	// Room needs p1 in Players to broadcast properly or just exists?
	// handleChat calls room.Broadcast.
	// We need to ensure room.Players has p1 if we want p1 to receive it?
	// Or we just verify no error.

	// Expectations
	mockServer.On("GetChatLimiter").Return(mockLimiter)
	mockServer.On("GetRoomManager").Return(mockRM)

	mockClient.On("GetID").Return("p1")
	mockClient.On("GetName").Return("Player1")
	// For room chat, GetRoom is called 3 times:
	// 1. In handleChat to get roomCode
	// 2. In handleChat to get roomInterface from manager
	// 3. Possibly logging?
	mockClient.On("GetRoom").Return("123")

	mockLimiter.On("AllowChat", "p1").Return(true, "")

	// Return the REAL *game.Room
	mockRM.On("GetRoom", "123").Return(room)

	// NOTE: room.Broadcast calls client.SendMessage for each player in room.
	// Since room has no players (NewMockRoom(..., nil)), nothing happens.
	// But we want to verify room logic was executed (i.e. not fell back to broadcast).
	// If we add p1 to room, we expect SendMessage on p1.
	// But p1 is mockClient.

	// Let's add p1 to room.
	room.Players["p1"] = &game.RoomPlayer{
		Client: mockClient,
		Seat:   0,
		Ready:  true,
	}

	// Expect p1 (mockClient) to receive the chat message echoed back
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgChat
	})).Return()

	// 3. Execution
	payload := protocol.ChatPayload{
		Content: "Hello Room",
		Scope:   "room",
	}
	msg := codec.MustNewMessage(protocol.MsgChat, payload)

	h.handleChat(mockClient, msg)

	// 4. Verification
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockLimiter.AssertExpectations(t)
	mockRM.AssertExpectations(t)
}

package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	r "github.com/palemoky/fight-the-landlord/internal/game/room"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

func TestHandler_HandleChat_Lobby(t *testing.T) {
	// 1. Setup
	mockServer := new(testutil.MockServer)
	mockClient := new(testutil.MockClient)
	mockLimiter := new(testutil.MockChatLimiter)

	h := NewHandler(HandlerDeps{
		Server:      mockServer,
		ChatLimiter: mockLimiter,
	})

	// 2. Expectations
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
	mockServer := new(testutil.MockServer)
	mockClient := new(testutil.MockClient)
	mockLimiter := new(testutil.MockChatLimiter)

	h := NewHandler(HandlerDeps{
		Server:      mockServer,
		ChatLimiter: mockLimiter,
	})

	// 2. Expectations
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
	mockServer := new(testutil.MockServer)
	mockClient := new(testutil.MockClient)
	mockLimiter := new(testutil.MockChatLimiter)

	// Use NewMockRoom helper which returns a real *r.Room
	room := r.NewMockRoom("123", nil)

	// Create a real RoomManager and add the room
	rm := r.NewRoomManager(nil, 10*time.Minute)
	rm.AddRoomForTest(room)

	h := NewHandler(HandlerDeps{
		Server:      mockServer,
		ChatLimiter: mockLimiter,
		RoomManager: rm,
	})

	// Expectations
	mockClient.On("GetID").Return("p1")
	mockClient.On("GetName").Return("Player1")
	mockClient.On("GetRoom").Return("123")
	mockLimiter.On("AllowChat", "p1").Return(true, "")

	// Add p1 to room so broadcast works
	room.Players["p1"] = &r.RoomPlayer{
		Client: mockClient,
		Seat:   0,
		Ready:  true,
	}

	// Expect p1 (mockClient) to receive the chat message
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgChat
	})).Return()

	// 2. Execution
	payload := protocol.ChatPayload{
		Content: "Hello Room",
		Scope:   "room",
	}
	msg := codec.MustNewMessage(protocol.MsgChat, payload)

	h.handleChat(mockClient, msg)

	// 3. Verification
	mockClient.AssertExpectations(t)
	mockLimiter.AssertExpectations(t)
}

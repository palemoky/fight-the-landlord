package handlers

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

func TestHandler_HandleCreateRoom_Maintenance(t *testing.T) {
	// Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	h := NewHandler(mockServer)

	// Expectations
	mockServer.On("IsMaintenanceMode").Return(true)
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgError
	})).Return()

	// Execute
	h.handleCreateRoom(mockClient)

	// Verify
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestHandler_HandleCreateRoom_Success(t *testing.T) {
	// Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockRoomManager := new(MockRoomManager)
	h := NewHandler(mockServer)

	// Expectations
	// Move GetID expectation up because NewMockRoom calls it
	mockClient.On("GetID").Return("p1")
	mockClient.On("GetName").Return("Player1")

	// The handler asserts the result is *game.Room, so we need to return that or a comparable struct
	// defined in a way the type assertion passed.
	// Since *game.Room is a concrete type in another package, we used NewMockRoom() helper in mock_test.go
	// which returns *game.Room
	mockRoom := NewMockRoom("123456", mockClient)

	mockServer.On("IsMaintenanceMode").Return(false)
	mockClient.On("GetRoom").Return("") // Not in room
	mockServer.On("GetRoomManager").Return(mockRoomManager)

	// mockRoomManager.CreateRoom returns (mockRoom, nil)
	mockRoomManager.On("CreateRoom", mockClient).Return(mockRoom, nil)

	// Expect SendMessage MsgRoomCreated
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgRoomCreated
	})).Return()

	// Execute
	h.handleCreateRoom(mockClient)

	// Verify
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockRoomManager.AssertExpectations(t)
}

func TestHandler_HandleJoinRoom_Success(t *testing.T) {
	// Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockRoomManager := new(MockRoomManager)
	h := NewHandler(mockServer)

	// Expectations
	// Move GetID expectation up
	mockClient.On("GetID").Return("p1")
	mockClient.On("GetName").Return("Player1")

	mockRoom := NewMockRoom("123456", mockClient)

	mockServer.On("IsMaintenanceMode").Return(false)
	mockClient.On("GetRoom").Return("")
	mockServer.On("GetRoomManager").Return(mockRoomManager)

	// JoinRoom params
	roomCode := "123456"
	mockRoomManager.On("JoinRoom", mockClient, roomCode).Return(mockRoom, nil)

	// Expect Success Message
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgRoomJoined
	})).Return()

	payload := protocol.JoinRoomPayload{RoomCode: roomCode}
	msg := encoding.MustNewMessage(protocol.MsgJoinRoom, payload)

	// Execute
	h.handleJoinRoom(mockClient, msg)

	// Verify
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockRoomManager.AssertExpectations(t)
}

func TestHandler_HandleQuickMatch_Success(t *testing.T) {
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockMatcher := new(MockMatcher)

	// There is no dependency on client in NewMockRoom here or room manager, just matcher.
	h := NewHandler(mockServer)

	// Expectations
	mockServer.On("IsMaintenanceMode").Return(false)
	mockClient.On("GetRoom").Return("") // Not in room

	mockServer.On("GetMatcher").Return(mockMatcher)
	mockMatcher.On("AddToQueue", mockClient).Return()

	h.handleQuickMatch(mockClient)

	mockServer.AssertExpectations(t)
	mockMatcher.AssertExpectations(t)
}

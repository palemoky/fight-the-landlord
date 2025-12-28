package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/network/client"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

func TestNewLobbyModel(t *testing.T) {
	// Setup dependencies
	mockClient := &client.Client{}
	inputModel := textinput.New()

	// Execute
	model := NewLobbyModel(mockClient, &inputModel)

	// Verify
	assert.NotNil(t, model)
	assert.Equal(t, mockClient, model.client)
	assert.Equal(t, &inputModel, model.input)

	// Verify Chat Input properties
	assert.Equal(t, "按 / 键聊天...", model.chatInput.Placeholder)
	assert.Equal(t, 50, model.chatInput.CharLimit)
	assert.Equal(t, 45, model.chatInput.Width, "Chat input width should be set to fit the chat box")
}

func TestLobbyModel_Navigation_Menu(t *testing.T) {
	model := &LobbyModel{
		selectedIndex: 0,
	}

	// Default menu items count is 6 (indices 0-5)
	// 0: Quick Match, 1: Create Room, 2: Join Room, 3: Leaderboard, 4: My Stats, 5: Rules

	// Test Down Key (Next)
	// 0 -> 1
	model.handleDownKey(PhaseLobby)
	assert.Equal(t, 1, model.selectedIndex)

	// Test wrapping around
	model.selectedIndex = 5
	model.handleDownKey(PhaseLobby)
	assert.Equal(t, 0, model.selectedIndex)

	// Test Up Key (Prev)
	// 0 -> 5 (Wrap around)
	model.handleUpKey(PhaseLobby)
	assert.Equal(t, 5, model.selectedIndex)

	// 5 -> 4
	model.handleUpKey(PhaseLobby)
	assert.Equal(t, 4, model.selectedIndex)
}

func TestLobbyModel_Navigation_RoomList(t *testing.T) {
	// Setup model with some mock rooms
	rooms := []protocol.RoomListItem{
		{RoomCode: "111", PlayerCount: 1},
		{RoomCode: "222", PlayerCount: 2},
		{RoomCode: "333", PlayerCount: 3},
	}

	model := &LobbyModel{
		availableRooms:  rooms,
		selectedRoomIdx: 0,
	}

	// Test Down Key (Next Room)
	// 0 -> 1
	model.handleDownKey(PhaseRoomList)
	assert.Equal(t, 1, model.selectedRoomIdx)

	// Test wrapping
	model.selectedRoomIdx = 2
	model.handleDownKey(PhaseRoomList)
	assert.Equal(t, 0, model.selectedRoomIdx)

	// Test Up Key (Prev Room)
	// 0 -> 2 (Wrap)
	model.handleUpKey(PhaseRoomList)
	assert.Equal(t, 2, model.selectedRoomIdx)

	// 2 -> 1
	model.handleUpKey(PhaseRoomList)
	assert.Equal(t, 1, model.selectedRoomIdx)
}

func TestLobbyModel_Navigation_EmptyRoomList(t *testing.T) {
	model := &LobbyModel{
		availableRooms:  []protocol.RoomListItem{},
		selectedRoomIdx: 0,
	}

	// Should not panic and indices shouldn't change if list is empty,
	// or logic prevents movement.
	// Based on implementation: if len > 0 checks are performed.

	model.handleDownKey(PhaseRoomList)
	assert.Equal(t, 0, model.selectedRoomIdx)

	model.handleUpKey(PhaseRoomList)
	assert.Equal(t, 0, model.selectedRoomIdx)
}

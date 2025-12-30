package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/server/storage"
)

func TestHandler_HandleGetStats_Success(t *testing.T) {
	// Setup
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLeaderboard := new(MockLeaderboard)
	h := NewHandler(mockServer)

	// Mock Data
	playerID := "p1"
	stats := &storage.PlayerStats{
		PlayerID:   playerID,
		PlayerName: "Player1",
		TotalGames: 10,
		Wins:       5,
		Score:      100,
	}

	// Expectations
	mockClient.On("GetID").Return(playerID)
	// mockClient.On("GetName").Return("Player1") // Used inside if stats==nil logic, but here stats != nil

	mockServer.On("GetLeaderboard").Return(mockLeaderboard)

	// GetPlayerStats returns interface{}, error
	mockLeaderboard.On("GetPlayerStats", mock.Anything, playerID).Return(stats, nil)
	// GetPlayerRank returns int64, error
	mockLeaderboard.On("GetPlayerRank", mock.Anything, playerID).Return(int64(1), nil)

	// Expect MsgStatsResult
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgStatsResult
	})).Return()

	// Execute
	h.handleGetStats(mockClient)

	// Verify
	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockLeaderboard.AssertExpectations(t)
}

func TestHandler_HandleGetStats_Empty(t *testing.T) {
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLeaderboard := new(MockLeaderboard)
	h := NewHandler(mockServer)

	playerID := "p1"
	playerName := "Player1"

	mockClient.On("GetID").Return(playerID)
	mockServer.On("GetLeaderboard").Return(mockLeaderboard)

	// Return nil, nil (no stats found)
	// Note: mock.Anything for context might match, but explicit nil for stats
	mockLeaderboard.On("GetPlayerStats", mock.Anything, playerID).Return(nil, nil)

	// When stats is nil, it calls client.GetName() to fill response
	mockClient.On("GetName").Return(playerName)

	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgStatsResult
	})).Return()

	h.handleGetStats(mockClient)

	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockLeaderboard.AssertExpectations(t)
}

func TestHandler_HandleGetStats_Error(t *testing.T) {
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLeaderboard := new(MockLeaderboard)
	h := NewHandler(mockServer)

	mockClient.On("GetID").Return("p1")
	mockServer.On("GetLeaderboard").Return(mockLeaderboard)

	mockLeaderboard.On("GetPlayerStats", mock.Anything, "p1").Return(nil, errors.New("db error"))

	// Expect Error Message
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgError
	})).Return()

	h.handleGetStats(mockClient)

	mockServer.AssertExpectations(t)
}

func TestHandler_HandleGetLeaderboard_Success(t *testing.T) {
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockLeaderboard := new(MockLeaderboard)
	h := NewHandler(mockServer)

	// Mock Entries
	entries := []interface{}{
		&storage.LeaderboardEntry{
			Rank:       1,
			PlayerID:   "p1",
			PlayerName: "Player1",
			Score:      100,
		},
	}

	mockServer.On("GetLeaderboard").Return(mockLeaderboard)

	// Default limit 10
	mockLeaderboard.On("GetLeaderboard", mock.Anything, 10).Return(entries, nil)

	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgLeaderboardResult
	})).Return()

	// Default payload
	// If msg payload is nil or parse fails, it defaults.
	// We send an empty message or valid one.
	// handleGetLeaderboard parses payload. If we send nil payload in message, ParsePayload might error?
	// encoding.ParsePayload handles it.

	// Creating a message with empty payload to trigger default logic
	msg := &protocol.Message{Type: protocol.MsgGetLeaderboard, Payload: []byte("{}")}

	h.handleGetLeaderboard(mockClient, msg)

	mockServer.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockLeaderboard.AssertExpectations(t)
}

func TestHandler_HandleGetOnlineCount(t *testing.T) {
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	h := NewHandler(mockServer)

	mockServer.On("GetOnlineCount").Return(42)
	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgOnlineCount
	})).Return()

	h.handleGetOnlineCount(mockClient)

	mockServer.AssertExpectations(t)
}

func TestHandler_HandleGetRoomList(t *testing.T) {
	mockServer := new(MockServer)
	mockClient := new(MockClient)
	mockRoomManager := new(MockRoomManager)
	h := NewHandler(mockServer)

	mockServer.On("GetRoomManager").Return(mockRoomManager)

	// RoomList returns []any
	rooms := []any{
		protocol.RoomListItem{RoomCode: "123", PlayerCount: 1},
	}
	mockRoomManager.On("GetRoomList").Return(rooms)

	mockClient.On("SendMessage", mock.MatchedBy(func(msg *protocol.Message) bool {
		return msg.Type == protocol.MsgRoomListResult
	})).Return()

	h.handleGetRoomList(mockClient)

	mockServer.AssertExpectations(t)
}

package handlers

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// --- MockClient ---

type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockClient) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockClient) GetRoom() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockClient) SetRoom(roomCode string) {
	m.Called(roomCode)
}

func (m *MockClient) SendMessage(msg *protocol.Message) {
	m.Called(msg)
}

func (m *MockClient) Close() {
	m.Called()
}

// --- MockServer ---

type MockServer struct {
	mock.Mock
}

func (m *MockServer) GetRedisStore() types.RedisStoreInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.RedisStoreInterface)
}

func (m *MockServer) GetLeaderboard() types.LeaderboardInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.LeaderboardInterface)
}

func (m *MockServer) GetSessionManager() types.SessionManagerInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.SessionManagerInterface)
}

func (m *MockServer) GetRoomManager() types.RoomManagerInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.RoomManagerInterface)
}

func (m *MockServer) GetMatcher() types.MatcherInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.MatcherInterface)
}

func (m *MockServer) IsMaintenanceMode() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockServer) GetOnlineCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockServer) Broadcast(msg *protocol.Message) {
	m.Called(msg)
}

func (m *MockServer) GetChatLimiter() types.ChatLimiterInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.ChatLimiterInterface)
}

func (m *MockServer) GetClientByID(id string) types.ClientInterface {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(types.ClientInterface)
}

func (m *MockServer) RegisterClient(id string, client types.ClientInterface) {
	m.Called(id, client)
}

func (m *MockServer) UnregisterClient(id string) {
	m.Called(id)
}

// --- MockChatLimiter ---

type MockChatLimiter struct {
	mock.Mock
}

func (m *MockChatLimiter) AllowChat(playerID string) (bool, string) {
	args := m.Called(playerID)
	return args.Bool(0), args.String(1)
}

// --- MockRoomManager ---

type MockRoomManager struct {
	mock.Mock
}

func (m *MockRoomManager) LeaveRoom(client types.ClientInterface) {
	m.Called(client)
}

func (m *MockRoomManager) CreateRoom(client types.ClientInterface) (any, error) {
	args := m.Called(client)
	return args.Get(0), args.Error(1)
}

func (m *MockRoomManager) JoinRoom(client types.ClientInterface, code string) (any, error) {
	args := m.Called(client, code)
	return args.Get(0), args.Error(1)
}

func (m *MockRoomManager) SetPlayerReady(client types.ClientInterface, ready bool) error {
	args := m.Called(client, ready)
	return args.Error(0)
}

func (m *MockRoomManager) GetRoom(code string) any {
	args := m.Called(code)
	return args.Get(0)
}

func (m *MockRoomManager) GetRoomList() []any {
	args := m.Called()
	return args.Get(0).([]any)
}

func (m *MockRoomManager) GetRoomByPlayerID(playerID string) any {
	args := m.Called(playerID)
	return args.Get(0)
}

func (m *MockRoomManager) GetActiveGamesCount() int {
	args := m.Called()
	return args.Int(0)
}

// --- MockMatcher ---

type MockMatcher struct {
	mock.Mock
}

func (m *MockMatcher) AddToQueue(client types.ClientInterface) {
	m.Called(client)
}

// --- MockLeaderboard ---
type MockLeaderboard struct {
	mock.Mock
}

func (m *MockLeaderboard) RecordGameResult(ctx context.Context, playerID, playerName string, isWinner, isLandlord bool) error {
	args := m.Called(ctx, playerID, playerName, isWinner, isLandlord)
	return args.Error(0)
}

func (m *MockLeaderboard) GetPlayerStats(ctx context.Context, playerID string) (interface{}, error) {
	args := m.Called(ctx, playerID)
	return args.Get(0), args.Error(1)
}

func (m *MockLeaderboard) GetPlayerRank(ctx context.Context, playerID string) (int64, error) {
	args := m.Called(ctx, playerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLeaderboard) GetLeaderboard(ctx context.Context, limit int) ([]interface{}, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]interface{}), args.Error(1)
}

// --- MockRedisStore ---
type MockRedisStore struct {
	mock.Mock
}

func (m *MockRedisStore) SaveRoom(ctx context.Context, room any) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRedisStore) DeleteRoom(ctx context.Context, roomCode string) error {
	args := m.Called(ctx, roomCode)
	return args.Error(0)
}

// --- MockSessionManager ---
type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) IsOnline(playerID string) bool {
	args := m.Called(playerID)
	return args.Bool(0)
}

func (m *MockSessionManager) CanReconnect(token, playerID string) bool {
	args := m.Called(token, playerID)
	return args.Bool(0)
}

func (m *MockSessionManager) GetSession(playerID string) interface{} {
	args := m.Called(playerID)
	return args.Get(0)
}

func (m *MockSessionManager) SetOnline(playerID string) {
	m.Called(playerID)
}

// --- Helper Functions ---
// NewMockRoom creates a simple *game.Room for testing
// NewMockRoom creates a simple *game.Room for testing
func NewMockRoom(code string, client types.ClientInterface) *game.Room {
	room := &game.Room{
		Code:    code,
		Players: make(map[string]*game.RoomPlayer),
	}
	if client != nil {
		room.Players[client.GetID()] = &game.RoomPlayer{
			Client: client,
			Seat:   0,
			Ready:  false,
		}
	}
	return room
}

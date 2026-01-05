//go:build !production

package testutil

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/server/storage"
)

// MockLeaderboard 排行榜 mock
type MockLeaderboard struct {
	mock.Mock
}

func (m *MockLeaderboard) RecordGameResult(ctx context.Context, playerID, playerName string, isWinner, isLandlord bool) error {
	args := m.Called(ctx, playerID, playerName, isWinner, isLandlord)
	return args.Error(0)
}

func (m *MockLeaderboard) GetPlayerStats(ctx context.Context, playerID string) (*storage.PlayerStats, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.PlayerStats), args.Error(1)
}

func (m *MockLeaderboard) GetPlayerRank(ctx context.Context, playerID string) (int64, error) {
	args := m.Called(ctx, playerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLeaderboard) GetLeaderboard(ctx context.Context, limit int) ([]*storage.LeaderboardEntry, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.LeaderboardEntry), args.Error(1)
}

// MockRedisStore Redis 存储 mock
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

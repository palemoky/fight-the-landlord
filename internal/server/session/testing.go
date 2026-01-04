//go:build !production

package session

import (
	"github.com/stretchr/testify/mock"
)

// MockGameSessionStore 游戏会话存储 mock
type MockGameSessionStore struct {
	mock.Mock
}

func (m *MockGameSessionStore) SaveGameSession(roomCode string, data any) error {
	args := m.Called(roomCode, data)
	return args.Error(0)
}

func (m *MockGameSessionStore) LoadGameSession(roomCode string) (any, error) {
	args := m.Called(roomCode)
	return args.Get(0), args.Error(1)
}

func (m *MockGameSessionStore) DeleteGameSession(roomCode string) error {
	args := m.Called(roomCode)
	return args.Error(0)
}

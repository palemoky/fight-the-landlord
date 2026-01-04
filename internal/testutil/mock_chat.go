//go:build !production

package testutil

import "github.com/stretchr/testify/mock"

// MockChatLimiter 聊天限制器 mock
type MockChatLimiter struct {
	mock.Mock
}

func (m *MockChatLimiter) AllowChat(playerID string) (allowed bool, reason string) {
	args := m.Called(playerID)
	return args.Bool(0), args.String(1)
}

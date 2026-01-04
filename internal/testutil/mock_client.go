//go:build !production

package testutil

import (
	"github.com/stretchr/testify/mock"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

// MockClient 实现 types.ClientInterface 的 mock
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

// SimpleClient 简单的 mock 客户端，不使用 testify（用于不需要断言的测试）
type SimpleClient struct {
	ID       string
	Name     string
	RoomCode string
	Messages []*protocol.Message
}

func (m *SimpleClient) GetID() string                     { return m.ID }
func (m *SimpleClient) GetName() string                   { return m.Name }
func (m *SimpleClient) GetRoom() string                   { return m.RoomCode }
func (m *SimpleClient) SetRoom(code string)               { m.RoomCode = code }
func (m *SimpleClient) SendMessage(msg *protocol.Message) { m.Messages = append(m.Messages, msg) }
func (m *SimpleClient) Close()                            {}

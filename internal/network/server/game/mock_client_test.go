package game

import "github.com/palemoky/fight-the-landlord/internal/network/protocol"

type MockClient struct {
	ID       string
	Name     string
	RoomCode string
	Messages []*protocol.Message
}

func (m *MockClient) GetID() string {
	return m.ID
}

func (m *MockClient) GetName() string {
	return m.Name
}

func (m *MockClient) GetRoom() string {
	return m.RoomCode
}

func (m *MockClient) SetRoom(roomCode string) {
	m.RoomCode = roomCode
}

func (m *MockClient) SendMessage(msg *protocol.Message) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockClient) Close() {
	// No-op for mock
}

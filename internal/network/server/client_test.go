package server

import (
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	// 模拟 Server
	server := &Server{}
	// 模拟 Conn (这里只能用 nil 替代，因为 websocket.Conn 很难 mock，
	// 真正的连接测试通常在集成测试中做，或者使用 httptest 启动真实 server)
	var conn *websocket.Conn

	client := NewClient(server, conn)

	assert.NotEmpty(t, client.ID)
	assert.NotEmpty(t, client.Name)
	assert.Equal(t, server, client.server)
	assert.NotNil(t, client.send)
}

func TestClient_SetGetRoom(t *testing.T) {
	t.Parallel()

	client := &Client{}

	tests := []struct {
		name     string
		roomID   string
		expected string
	}{
		{"Set room A", "room-a", "room-a"},
		{"Set empty room", "", ""},
		{"Set room B", "room-b", "room-b"},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			client.SetRoom(tt.roomID)
			assert.Equal(t, tt.expected, client.GetRoom())
		})
	}
}

func TestClient_SetGetRoom_Concurrency(t *testing.T) {
	t.Parallel()

	client := &Client{}
	var wg sync.WaitGroup
	count := 100

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			client.SetRoom("room-concurrent")
			_ = client.GetRoom()
		}()
	}

	wg.Wait()
	assert.Equal(t, "room-concurrent", client.GetRoom())
}

func TestClient_Close(t *testing.T) {
	t.Parallel()

	client := &Client{
		send: make(chan []byte, 1),
	}

	// First close
	client.Close()
	assert.True(t, client.closed)

	// Second close (should be safe)
	assert.NotPanics(t, func() {
		client.Close()
	})

	// Check channel closed
	_, ok := <-client.send
	assert.False(t, ok)
}

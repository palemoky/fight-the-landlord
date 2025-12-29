package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

var upgrader = websocket.Upgrader{}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		// simple echo
		_ = c.WriteMessage(mt, message)
	}
}

func TestClient_ConnectAndSend(t *testing.T) {
	// Start a mock WS server that echoes messages
	s := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer s.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	client := NewClient(wsURL)
	assert.NotNil(t, client)

	// Connect
	err := client.Connect()
	assert.NoError(t, err)
	defer client.Close()

	// Wait for connection to establish
	time.Sleep(100 * time.Millisecond)
	assert.True(t, client.IsConnected())

	// Send a message
	msg := encoding.MustNewMessage(protocol.MsgPing, protocol.PingPayload{Timestamp: 123456})
	err = client.SendMessage(msg)
	assert.NoError(t, err)

	// Receive echo (blocks until message)
	// Since we are using an echo server, we expect to get the binary message back.
	// But client.readPump logic tries to Decode it.
	// The echo server sends back exactly what we sent (BinaryMessage).
	// So protocol.Decode should work if it's a valid message.

	receivedMsg, err := client.ReceiveWithTimeout(1 * time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, receivedMsg)
	assert.Equal(t, protocol.MsgPing, receivedMsg.Type)
}

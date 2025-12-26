package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionManager_CRUD(t *testing.T) {
	sm := NewSessionManager()

	// Create
	session := sm.CreateSession("p1", "Player1")
	assert.NotNil(t, session)
	assert.Equal(t, "p1", session.PlayerID)
	assert.Equal(t, "Player1", session.PlayerName)
	assert.NotEmpty(t, session.ReconnectToken)
	assert.True(t, session.IsOnline)

	// Get by ID
	s1 := sm.GetSession("p1")
	assert.Equal(t, session, s1)

	// Get by Token
	s2 := sm.GetSessionByToken(session.ReconnectToken)
	assert.Equal(t, session, s2)

	// Delete
	sm.DeleteSession("p1")
	assert.Nil(t, sm.GetSession("p1"))
	assert.Nil(t, sm.GetSessionByToken(session.ReconnectToken))
}

func TestSessionManager_OnlineStatus(t *testing.T) {
	sm := NewSessionManager()
	sm.CreateSession("p1", "Player1")

	// Set Offline
	sm.SetOffline("p1")
	s1 := sm.GetSession("p1")
	assert.False(t, s1.IsOnline)
	assert.False(t, s1.DisconnectedAt.IsZero())

	// Set Online
	sm.SetOnline("p1")
	s2 := sm.GetSession("p1")
	assert.True(t, s2.IsOnline)
	assert.True(t, s2.DisconnectedAt.IsZero())
}

func TestSessionManager_CanReconnect(t *testing.T) {
	sm := NewSessionManager()
	session := sm.CreateSession("p1", "Player1")
	token := session.ReconnectToken

	// Case 1: Online player cannot match "reconnect" conditions usually (or implementation allows it?)
	// Implementation: if !session.IsOnline && time.Since ...
	// Wait, Check implementation:
	// func (sm *SessionManager) CanReconnect(token, playerID string) bool {
	//      // ...
	// 		if !session.IsOnline && time.Since(session.DisconnectedAt) > reconnectTimeout { return false }
	//      return true
	// }
	// So if online, it returns true? Let's check logic.
	// If it returns true, it means token matches.
	assert.True(t, sm.CanReconnect(token, "p1"))

	// Case 2: Offline within limits
	sm.SetOffline("p1")
	assert.True(t, sm.CanReconnect(token, "p1"))

	// Case 3: Wrong token
	assert.False(t, sm.CanReconnect("wrong-token", "p1"))

	// Case 4: Wrong PlayerID
	assert.False(t, sm.CanReconnect(token, "p2"))

	// Case 5: Timed out
	// Manually hack DisconnectedAt to be in the past
	session.DisconnectedAt = time.Now().Add(-3 * time.Minute) // Timeout is 2 mins
	assert.False(t, sm.CanReconnect(token, "p1"))
}

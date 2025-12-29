package core

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
	s1 := sm.GetSession("p1").(*PlayerSession)
	assert.False(t, s1.IsOnline)
	assert.False(t, s1.DisconnectedAt.IsZero())

	// Set Online
	sm.SetOnline("p1")
	s2 := sm.GetSession("p1").(*PlayerSession)
	assert.True(t, s2.IsOnline)
	assert.True(t, s2.DisconnectedAt.IsZero())
}

func TestSessionManager_CanReconnect(t *testing.T) {
	sm := NewSessionManager()
	session := sm.CreateSession("p1", "Player1")
	validToken := session.ReconnectToken

	tests := []struct {
		name     string
		token    string
		playerID string
		setup    func()
		expected bool
	}{
		{
			name:     "Valid reconnection (online)",
			token:    validToken,
			playerID: "p1",
			setup:    func() {}, // Already online
			expected: true,
		},
		{
			name:     "Valid reconnection (offline)",
			token:    validToken,
			playerID: "p1",
			setup: func() {
				sm.SetOffline("p1")
			},
			expected: true,
		},
		{
			name:     "Invalid token",
			token:    "wrong-token",
			playerID: "p1",
			setup:    func() {},
			expected: false,
		},
		{
			name:     "Wrong player ID",
			token:    validToken,
			playerID: "p2",
			setup:    func() {},
			expected: false,
		},
		{
			name:     "Expired session",
			token:    validToken,
			playerID: "p1",
			setup: func() {
				sm.SetOffline("p1")
				// Hack internal time for testing
				session.mu.Lock()
				session.DisconnectedAt = time.Now().Add(-3 * time.Minute)
				session.mu.Unlock()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.setup != nil {
				tt.setup()
			}
			assert.Equal(t, tt.expected, sm.CanReconnect(tt.token, tt.playerID))
		})
	}
}

package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionManager_CRUD(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()

	t.Run("Valid reconnection (online)", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		session := sm.CreateSession("p1", "Player1")
		assert.True(t, sm.CanReconnect(session.ReconnectToken, "p1"))
	})

	t.Run("Valid reconnection (offline)", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		session := sm.CreateSession("p1", "Player1")
		sm.SetOffline("p1")
		assert.True(t, sm.CanReconnect(session.ReconnectToken, "p1"))
	})

	t.Run("Invalid token", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		sm.CreateSession("p1", "Player1")
		assert.False(t, sm.CanReconnect("wrong-token", "p1"))
	})

	t.Run("Wrong player ID", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		session := sm.CreateSession("p1", "Player1")
		assert.False(t, sm.CanReconnect(session.ReconnectToken, "p2"))
	})

	t.Run("Expired session", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		session := sm.CreateSession("p1", "Player1")
		sm.SetOffline("p1")
		// Hack internal time for testing
		session.mu.Lock()
		session.DisconnectedAt = time.Now().Add(-3 * time.Minute)
		session.mu.Unlock()
		assert.False(t, sm.CanReconnect(session.ReconnectToken, "p1"))
	})
}

func TestSessionManager_SetRoom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		playerID     string
		roomCode     string
		shouldCreate bool
	}{
		{"set room for existing player", "p1", "123456", true},
		{"set room for non-existent player", "p999", "123456", false},
		{"clear room", "p1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sm := NewSessionManager()
			if tt.shouldCreate {
				sm.CreateSession("p1", "Player1")
			}

			sm.SetRoom(tt.playerID, tt.roomCode)

			if tt.shouldCreate && tt.playerID == "p1" {
				session := sm.GetSession("p1").(*PlayerSession)
				assert.Equal(t, tt.roomCode, session.RoomCode)
			}
		})
	}
}

func TestSessionManager_IsOnline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(sm *SessionManager)
		playerID   string
		wantOnline bool
	}{
		{
			name: "online player",
			setup: func(sm *SessionManager) {
				sm.CreateSession("p1", "Player1")
			},
			playerID:   "p1",
			wantOnline: true,
		},
		{
			name: "offline player",
			setup: func(sm *SessionManager) {
				sm.CreateSession("p1", "Player1")
				sm.SetOffline("p1")
			},
			playerID:   "p1",
			wantOnline: false,
		},
		{
			name:       "non-existent player",
			setup:      func(_ *SessionManager) {},
			playerID:   "p999",
			wantOnline: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sm := NewSessionManager()
			tt.setup(sm)
			assert.Equal(t, tt.wantOnline, sm.IsOnline(tt.playerID))
		})
	}
}

func TestSessionManager_GetSessionByToken_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid token returns nil", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		sm.CreateSession("p1", "Player1")
		assert.Nil(t, sm.GetSessionByToken("invalid-token"))
	})

	t.Run("empty token returns nil", func(t *testing.T) {
		t.Parallel()
		sm := NewSessionManager()
		sm.CreateSession("p1", "Player1")
		assert.Nil(t, sm.GetSessionByToken(""))
	})
}

func TestSessionManager_SetOffline_NonExistent(t *testing.T) {
	t.Parallel()
	sm := NewSessionManager()
	// Should not panic
	sm.SetOffline("non-existent")
}

func TestSessionManager_SetOnline_NonExistent(t *testing.T) {
	t.Parallel()
	sm := NewSessionManager()
	// Should not panic
	sm.SetOnline("non-existent")
}

func TestSessionManager_DeleteSession_NonExistent(t *testing.T) {
	t.Parallel()
	sm := NewSessionManager()
	// Should not panic
	sm.DeleteSession("non-existent")
}

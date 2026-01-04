package room

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/apperrors"
	"github.com/palemoky/fight-the-landlord/internal/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

func TestNotifyPlayerOffline_AllPlayersOffline(t *testing.T) {
	t.Parallel()

	// Setup
	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	client1 := testutil.NewSimpleClient("p1", "Player1")
	client2 := testutil.NewSimpleClient("p2", "Player2")
	client3 := testutil.NewSimpleClient("p3", "Player3")

	// Create room with 3 players
	room, err := rm.CreateRoom(client1)
	require.NoError(t, err)
	_, err = rm.JoinRoom(client2, room.Code)
	require.NoError(t, err)
	_, err = rm.JoinRoom(client3, room.Code)
	require.NoError(t, err)

	// All players go offline
	rm.NotifyPlayerOffline(client1)
	rm.NotifyPlayerOffline(client2)
	rm.NotifyPlayerOffline(client3)

	// Room should be deleted
	assert.Nil(t, rm.GetRoom(room.Code))
}

func TestNotifyPlayerOffline_PartialOffline(t *testing.T) {
	t.Parallel()

	// Setup
	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	client1 := testutil.NewSimpleClient("p1", "Player1")
	client2 := testutil.NewSimpleClient("p2", "Player2")
	client3 := testutil.NewSimpleClient("p3", "Player3")

	// Create room with 3 players
	room, err := rm.CreateRoom(client1)
	require.NoError(t, err)
	_, err = rm.JoinRoom(client2, room.Code)
	require.NoError(t, err)
	_, err = rm.JoinRoom(client3, room.Code)
	require.NoError(t, err)

	// Only one player goes offline
	rm.NotifyPlayerOffline(client1)

	// Room should still exist
	assert.NotNil(t, rm.GetRoom(room.Code))

	// Verify offline notification was sent to other players
	assert.Eventually(t, func() bool {
		return len(client2.SentMessages()) > 0 || len(client3.SentMessages()) > 0
	}, time.Second, 10*time.Millisecond)
}

func TestNotifyPlayerOffline_NotInRoom(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	client := testutil.NewSimpleClient("p1", "Player1")

	// Client not in any room - should not panic
	assert.NotPanics(t, func() {
		rm.NotifyPlayerOffline(client)
	})
}

func TestReconnectPlayer_Success(t *testing.T) {
	t.Parallel()

	// Setup
	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	oldClient := testutil.NewSimpleClient("p1", "Player1")
	newClient := testutil.NewSimpleClient("p1", "Player1") // Same ID, new connection

	// Create room
	room, err := rm.CreateRoom(oldClient)
	require.NoError(t, err)

	// Reconnect
	err = rm.ReconnectPlayer(oldClient, newClient)
	require.NoError(t, err)

	// Verify new client is in room
	assert.Equal(t, room.Code, newClient.GetRoom())

	// Verify room player reference updated
	rm.mu.RLock()
	r := rm.rooms[room.Code]
	rm.mu.RUnlock()

	r.mu.RLock()
	player := r.Players[newClient.GetID()]
	r.mu.RUnlock()

	assert.Equal(t, newClient, player.Client)
}

func TestReconnectPlayer_RoomNotFound(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	oldClient := testutil.NewSimpleClient("p1", "Player1")
	newClient := testutil.NewSimpleClient("p1", "Player1")

	// Set room code but room doesn't exist
	oldClient.SetRoom("NONEXISTENT")

	err := rm.ReconnectPlayer(oldClient, newClient)
	assert.ErrorIs(t, err, apperrors.ErrRoomNotFound)
}

func TestReconnectPlayer_PlayerNotInRoom(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	client1 := testutil.NewSimpleClient("p1", "Player1")
	oldClient := testutil.NewSimpleClient("p2", "Player2")
	newClient := testutil.NewSimpleClient("p2", "Player2")

	// Create room with client1
	room, err := rm.CreateRoom(client1)
	require.NoError(t, err)

	// Try to reconnect client2 who was never in the room
	oldClient.SetRoom(room.Code)
	err = rm.ReconnectPlayer(oldClient, newClient)
	assert.ErrorIs(t, err, apperrors.ErrNotInRoom)
}

func TestReconnectPlayer_NotInAnyRoom(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	oldClient := testutil.NewSimpleClient("p1", "Player1")
	newClient := testutil.NewSimpleClient("p1", "Player1")

	// Client not in any room
	err := rm.ReconnectPlayer(oldClient, newClient)
	assert.NoError(t, err) // Should return nil, not error
}

func TestGenerateRoomCode_Uniqueness(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)

	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := rm.generateRoomCode()
		assert.Len(t, code, roomCodeLength)
		assert.False(t, codes[code], "Generated duplicate room code: %s", code)
		codes[code] = true

		// Add to rooms to test collision avoidance
		rm.rooms[code] = &Room{Code: code}
	}
}

func TestCleanup_TimeoutRooms(t *testing.T) {
	t.Parallel()

	// Use short timeout for testing
	rm := NewRoomManager(storage.NewRedisStore(nil), 100*time.Millisecond)
	client := testutil.NewSimpleClient("p1", "Player1")

	// Create room
	room, err := rm.CreateRoom(client)
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	rm.cleanup()

	// Room should be deleted
	assert.Nil(t, rm.GetRoom(room.Code))

	// Client should be removed from room
	assert.Empty(t, client.GetRoom())
}

func TestCleanup_DoesNotRemoveActiveRooms(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	client := testutil.NewSimpleClient("p1", "Player1")

	// Create room
	room, err := rm.CreateRoom(client)
	require.NoError(t, err)

	// Run cleanup immediately (room is fresh)
	rm.cleanup()

	// Room should still exist
	assert.NotNil(t, rm.GetRoom(room.Code))
}

func TestCleanup_DoesNotRemovePlayingRooms(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 100*time.Millisecond)
	client := testutil.NewSimpleClient("p1", "Player1")

	// Create room
	room, err := rm.CreateRoom(client)
	require.NoError(t, err)

	// Change state to playing
	room.mu.Lock()
	room.State = RoomStatePlaying
	room.mu.Unlock()

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	rm.cleanup()

	// Room should NOT be deleted (playing state)
	assert.NotNil(t, rm.GetRoom(room.Code))
}

func TestSetAllPlayersReady(t *testing.T) {
	t.Parallel()

	rm := NewRoomManager(storage.NewRedisStore(nil), 10*time.Minute)
	client1 := testutil.NewSimpleClient("p1", "Player1")
	client2 := testutil.NewSimpleClient("p2", "Player2")
	client3 := testutil.NewSimpleClient("p3", "Player3")

	// Create room with 3 players
	room, err := rm.CreateRoom(client1)
	require.NoError(t, err)
	_, err = rm.JoinRoom(client2, room.Code)
	require.NoError(t, err)
	_, err = rm.JoinRoom(client3, room.Code)
	require.NoError(t, err)

	// Initially not ready
	room.mu.RLock()
	for _, p := range room.Players {
		assert.False(t, p.Ready)
	}
	room.mu.RUnlock()

	// Set all ready
	room.SetAllPlayersReady()

	// Verify all ready
	room.mu.RLock()
	for _, p := range room.Players {
		assert.True(t, p.Ready)
	}
	room.mu.RUnlock()
}

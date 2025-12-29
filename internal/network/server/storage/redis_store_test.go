package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/network/server/core"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
)

func newTestRedisStore(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	store := NewRedisStore(client)
	return store, mr
}

func TestRedisStore_SaveLoadDeleteRoom(t *testing.T) {
	store, mr := newTestRedisStore(t)
	defer mr.Close()
	ctx := context.Background()

	// Create a game.Room manually
	// We need to ensure SerializeForRedis works.
	// If SerializeForRedis depends on internal state not initialized here, it might panic.
	// Assuming minimal initialization is enough or SerializeForRedis is robust.
	room := &game.Room{
		Code:      "123",
		State:     1,
		CreatedAt: time.Now(),
		Players:   make(map[string]*game.RoomPlayer),
	}
	// Add logic to make SerializeForRedis happy if needed (e.g. avoid nil pointers)
	// Looking at redis_store.go, it calls room.SerializeForRedis(func() { ... })
	// which calls the callback. The callback accesses room fields.

	// Save
	err := store.SaveRoom(ctx, room)
	assert.NoError(t, err)

	// Load
	loadedData, err := store.LoadRoom(ctx, "123")
	assert.NoError(t, err)
	assert.NotNil(t, loadedData)
	assert.Equal(t, "123", loadedData.Code)
	assert.Equal(t, 1, loadedData.State)

	// Delete
	err = store.DeleteRoom(ctx, "123")
	assert.NoError(t, err)

	// Verify Delete
	loadedData, err = store.LoadRoom(ctx, "123")
	assert.NoError(t, err)
	assert.Nil(t, loadedData)
}

func TestRedisStore_MatchQueue(t *testing.T) {
	store, mr := newTestRedisStore(t)
	defer mr.Close()
	ctx := context.Background()

	// Add
	err := store.AddToMatchQueue(ctx, "p1")
	assert.NoError(t, err)
	err = store.AddToMatchQueue(ctx, "p2")
	assert.NoError(t, err)

	// Length
	n, err := store.GetMatchQueueLength(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), n)

	// Remove
	err = store.RemoveFromMatchQueue(ctx, "p1")
	assert.NoError(t, err)

	n, err = store.GetMatchQueueLength(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)

	// Pop
	result, err := store.PopFromMatchQueue(ctx, 2) // Request 2, but only 1 left
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "p2", result[0])
}

func TestRedisStore_Session(t *testing.T) {
	store, mr := newTestRedisStore(t)
	defer mr.Close()
	ctx := context.Background()

	session := &core.PlayerSession{
		PlayerID:       "p1",
		PlayerName:     "Player1",
		ReconnectToken: "token123",
		RoomCode:       "room1",
		IsOnline:       true,
	}

	// Save
	err := store.SaveSession(ctx, session)
	assert.NoError(t, err)

	// Load
	loaded, err := store.LoadSession(ctx, "p1")
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.PlayerName, loaded.PlayerName)
	assert.Equal(t, session.IsOnline, loaded.IsOnline)

	// Delete
	err = store.DeleteSession(ctx, "p1")
	assert.NoError(t, err)

	// Verify Delete
	loaded, err = store.LoadSession(ctx, "p1")
	assert.NoError(t, err)
	assert.Nil(t, loaded)
}

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestRedisStore(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	t.Helper()
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

	// Create test room data
	roomData := &RoomData{
		Code:        "TEST123",
		State:       1,
		Players:     []PlayerData{},
		PlayerOrder: []string{},
		CreatedAt:   time.Now().Unix(),
	}

	// Save
	err := store.SaveRoom(ctx, roomData.Code, roomData)
	assert.NoError(t, err)

	// Load
	loadedData, err := store.LoadRoom(ctx, roomData.Code)
	assert.NoError(t, err)
	assert.NotNil(t, loadedData)
	assert.Equal(t, roomData.Code, loadedData.Code)
	assert.Equal(t, roomData.State, loadedData.State)

	// Delete
	err = store.DeleteRoom(ctx, roomData.Code)
	assert.NoError(t, err)

	// Verify Delete
	loadedData, err = store.LoadRoom(ctx, roomData.Code)
	assert.NoError(t, err)
	assert.Nil(t, loadedData)
}

func TestRedisStore_MatchQueue(t *testing.T) {
	t.Parallel()

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

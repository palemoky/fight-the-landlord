package room

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/testutil"
)

func TestRoomManager_GetRoomList(t *testing.T) {
	t.Parallel()

	// Initialize RoomManager with nil server (ok for this test)
	rm := NewRoomManager(nil, 10*time.Minute)

	// Manually add a suitable room
	room := &Room{
		Code:        "123456",
		State:       RoomStateWaiting,
		Players:     make(map[string]*RoomPlayer),
		PlayerOrder: []string{},
		CreatedAt:   time.Now(),
	}
	// Add a dummy player
	room.Players["p1"] = &RoomPlayer{
		Client: &testutil.SimpleClient{ID: "p1", Name: "Player1"},
		Seat:   0,
	}

	rm.rooms["123456"] = room

	// Execute
	rooms := rm.GetRoomList()

	// Verify
	assert.Len(t, rooms, 1)
	roomItem := rooms[0]
	assert.Equal(t, "123456", roomItem.RoomCode)
	assert.Equal(t, 1, roomItem.PlayerCount)
	assert.Equal(t, 3, roomItem.MaxPlayers)
}

func TestRoom_CheckAllReady(t *testing.T) {
	t.Parallel()

	room := &Room{
		Players: make(map[string]*RoomPlayer),
	}

	// Case 1: Not enough players
	room.Players["p1"] = &RoomPlayer{Ready: true}
	room.Players["p2"] = &RoomPlayer{Ready: true}
	assert.False(t, room.checkAllReady())

	// Case 2: Enough players, but not all ready
	room.Players["p3"] = &RoomPlayer{Ready: false}
	assert.False(t, room.checkAllReady())

	// Case 3: All ready
	room.Players["p3"].Ready = true
	assert.True(t, room.checkAllReady())
}

func TestRoom_GetPlayerInfo(t *testing.T) {
	t.Parallel()

	room := &Room{
		Players: make(map[string]*RoomPlayer),
	}
	client := &testutil.SimpleClient{ID: "p1", Name: "TestPlayer"}
	room.Players["p1"] = &RoomPlayer{
		Client:     client,
		Seat:       1,
		Ready:      true,
		IsLandlord: false,
	}

	info := room.GetPlayerInfo("p1")

	assert.Equal(t, "p1", info.ID)
	assert.Equal(t, "TestPlayer", info.Name)
	assert.Equal(t, 1, info.Seat)
	assert.True(t, info.Ready)
}

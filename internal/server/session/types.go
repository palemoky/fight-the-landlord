package session

import (
	"github.com/palemoky/fight-the-landlord/internal/game/room"
)

// Type aliases for backward compatibility
type (
	RoomState = room.RoomState
)

// Re-export room state constants
const (
	RoomStateWaiting = room.RoomStateWaiting
	RoomStateReady   = room.RoomStateReady
	RoomStateBidding = room.RoomStateBidding
	RoomStatePlaying = room.RoomStatePlaying
	RoomStateEnded   = room.RoomStateEnded
)

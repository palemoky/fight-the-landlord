package session

import (
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// Type aliases for convenience within the session package
type (
	RoomState           = types.RoomState
	RoomInterface       = types.RoomInterface
	RoomPlayerInterface = types.RoomPlayerInterface
)

// Re-export constants for convenience
const (
	RoomStateWaiting = types.RoomStateWaiting
	RoomStateReady   = types.RoomStateReady
	RoomStateBidding = types.RoomStateBidding
	RoomStatePlaying = types.RoomStatePlaying
	RoomStateEnded   = types.RoomStateEnded
)

// PlayerData 用于初始化 GameSession 的玩家数据
type PlayerData struct {
	ID   string
	Name string
	Seat int
}

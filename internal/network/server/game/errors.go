package game

import (
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// RoomError type alias
type RoomError = types.RoomError

// Error variables
var (
	ErrRoomNotFound = &RoomError{Code: protocol.ErrCodeRoomNotFound, Message: "房间不存在"}
	ErrRoomFull     = &RoomError{Code: protocol.ErrCodeRoomFull, Message: "房间已满"}
	ErrNotInRoom    = &RoomError{Code: protocol.ErrCodeNotInRoom, Message: "您不在房间中"}
	ErrGameStarted  = &RoomError{Code: protocol.ErrCodeGameNotStart, Message: "游戏已开始"}
	ErrGameNotStart = &RoomError{Code: protocol.ErrCodeGameNotStart, Message: "游戏尚未开始"}
	ErrNotYourTurn  = &RoomError{Code: protocol.ErrCodeNotYourTurn, Message: "还没轮到您"}
	ErrInvalidCards = &RoomError{Code: protocol.ErrCodeInvalidCards, Message: "无效的牌型"}
	ErrCannotBeat   = &RoomError{Code: protocol.ErrCodeCannotBeat, Message: "您的牌大不过上家"}
	ErrMustPlay     = &RoomError{Code: protocol.ErrCodeMustPlay, Message: "您必须出牌"}
)

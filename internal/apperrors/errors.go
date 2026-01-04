package apperrors

import (
	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

// GameError 游戏错误（房间和会话共享）
type GameError struct {
	Code    int
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}

// 预定义错误
var (
	ErrRoomNotFound = &GameError{Code: protocol.ErrCodeRoomNotFound, Message: "房间不存在"}
	ErrRoomFull     = &GameError{Code: protocol.ErrCodeRoomFull, Message: "房间已满"}
	ErrNotInRoom    = &GameError{Code: protocol.ErrCodeNotInRoom, Message: "您不在房间中"}
	ErrGameStarted  = &GameError{Code: protocol.ErrCodeGameNotStart, Message: "游戏已开始"}
	ErrGameNotStart = &GameError{Code: protocol.ErrCodeGameNotStart, Message: "游戏尚未开始"}
	ErrNotYourTurn  = &GameError{Code: protocol.ErrCodeNotYourTurn, Message: "还没轮到您"}
	ErrInvalidCards = &GameError{Code: protocol.ErrCodeInvalidCards, Message: "无效的牌型"}
	ErrCannotBeat   = &GameError{Code: protocol.ErrCodeCannotBeat, Message: "您的牌大不过上家"}
	ErrMustPlay     = &GameError{Code: protocol.ErrCodeMustPlay, Message: "您必须出牌"}
)

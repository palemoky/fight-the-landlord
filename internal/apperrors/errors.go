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

// newGameError 创建游戏错误，从 protocol.ErrorMessages 获取消息
func newGameError(code int) *GameError {
	message := protocol.ErrorMessages[code]
	if message == "" {
		message = protocol.ErrorMessages[protocol.ErrCodeUnknown]
	}
	return &GameError{Code: code, Message: message}
}

// 预定义错误
var (
	ErrRoomNotFound = newGameError(protocol.ErrCodeRoomNotFound)
	ErrRoomFull     = newGameError(protocol.ErrCodeRoomFull)
	ErrNotInRoom    = newGameError(protocol.ErrCodeNotInRoom)
	ErrGameStarted  = newGameError(protocol.ErrCodeGameStarted)
	ErrGameNotStart = newGameError(protocol.ErrCodeGameNotStart)
	ErrNotYourTurn  = newGameError(protocol.ErrCodeNotYourTurn)
	ErrInvalidCards = newGameError(protocol.ErrCodeInvalidCards)
	ErrCannotBeat   = newGameError(protocol.ErrCodeCannotBeat)
	ErrMustPlay     = newGameError(protocol.ErrCodeMustPlay)
)

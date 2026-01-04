package protocol

// 错误码
const (
	ErrCodeUnknown           = 1000
	ErrCodeInvalidMsg        = 1001
	ErrCodeRateLimit         = 1002 // 速率限制
	ErrCodeRoomNotFound      = 2001
	ErrCodeRoomFull          = 2002
	ErrCodeNotInRoom         = 2003
	ErrCodeGameStarted       = 2004 // 游戏已开始
	ErrCodeGameNotStart      = 3001
	ErrCodeNotYourTurn       = 3002
	ErrCodeInvalidCards      = 3003
	ErrCodeCannotBeat        = 3004
	ErrCodeMustPlay          = 3005
	ErrCodeServerMaintenance = 5003 // 服务器维护中
)

// ErrorMessages 错误码对应的消息
var ErrorMessages = map[int]string{
	ErrCodeUnknown:           "未知错误",
	ErrCodeInvalidMsg:        "无效的消息格式",
	ErrCodeRateLimit:         "请求过于频繁",
	ErrCodeRoomNotFound:      "房间不存在",
	ErrCodeRoomFull:          "房间已满",
	ErrCodeNotInRoom:         "您不在房间中",
	ErrCodeGameStarted:       "游戏已开始",
	ErrCodeGameNotStart:      "游戏尚未开始",
	ErrCodeNotYourTurn:       "还没轮到您",
	ErrCodeInvalidCards:      "无效的牌型",
	ErrCodeCannotBeat:        "您的牌大不过上家",
	ErrCodeMustPlay:          "您必须出牌",
	ErrCodeServerMaintenance: "服务器维护中",
}

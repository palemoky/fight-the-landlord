package session

import (
	"sync"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
	"github.com/palemoky/fight-the-landlord/internal/rule"
)

// GameState 游戏状态
type GameState int

const (
	GameStateInit GameState = iota
	GameStateBidding
	GameStatePlaying
	GameStateEnded
)

// GamePlayer 游戏中的玩家
type GamePlayer struct {
	ID         string
	Name       string
	Seat       int
	Hand       []card.Card
	IsLandlord bool
	IsOffline  bool // 是否离线
}

// GameSession 游戏会话
type GameSession struct {
	room    RoomInterface
	state   GameState
	players []*GamePlayer // 按座位顺序

	deck        card.Deck
	bottomCards []card.Card

	// 叫地主相关
	currentBidder int // 当前叫地主的玩家索引
	highestBidder int // 叫地主的玩家索引，-1 表示没人叫
	bidCount      int // 叫地主轮数

	// 出牌相关
	currentPlayer     int             // 当前出牌玩家索引
	lastPlayedHand    rule.ParsedHand // 上家出牌
	lastPlayerIdx     int             // 上家索引
	consecutivePasses int             // 连续 PASS 次数

	// 超时控制
	turnTimer        *time.Timer
	offlineWaitTimer *time.Timer   // 离线等待计时器
	remainingTime    time.Duration // 暂停时剩余的时间
	timerStartTime   time.Time     // 计时器开始时间
	timerMu          sync.Mutex

	mu sync.RWMutex
}

// NewGameSession 创建游戏会话
func NewGameSession(room RoomInterface) *GameSession {
	playerOrder := room.GetPlayerOrder()
	players := make([]*GamePlayer, len(playerOrder))
	for i, id := range playerOrder {
		rp := room.GetPlayer(id)
		players[i] = &GamePlayer{
			ID:   id,
			Name: rp.GetClient().GetName(),
			Seat: i,
		}
	}

	return &GameSession{
		room:          room,
		state:         GameStateInit,
		players:       players,
		highestBidder: -1,
	}
}

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

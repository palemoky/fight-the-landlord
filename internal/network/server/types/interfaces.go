package types

import (
	"context"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// ServerContext 服务器上下文接口 - 避免循环依赖
type ServerContext interface {
	GetRedisStore() RedisStoreInterface
	GetLeaderboard() LeaderboardInterface
	GetSessionManager() SessionManagerInterface
	GetRoomManager() RoomManagerInterface
	GetMatcher() MatcherInterface
	IsMaintenanceMode() bool
	GetOnlineCount() int
	Broadcast(msg *protocol.Message)
	BroadcastToLobby(msg *protocol.Message)
	GetChatLimiter() ChatLimiterInterface
	GetClientByID(id string) ClientInterface
	RegisterClient(id string, client ClientInterface)
	UnregisterClient(id string)
}

// RedisStoreInterface Redis存储接口
type RedisStoreInterface interface {
	SaveRoom(ctx context.Context, room any) error
	DeleteRoom(ctx context.Context, roomCode string) error
}

// LeaderboardInterface 排行榜接口
type LeaderboardInterface interface {
	RecordGameResult(ctx context.Context, playerID, playerName string, isWinner, isLandlord bool) error
	GetPlayerStats(ctx context.Context, playerID string) (interface{}, error)
	GetPlayerRank(ctx context.Context, playerID string) (int64, error)
	GetLeaderboard(ctx context.Context, limit int) ([]interface{}, error)
}

// SessionManagerInterface 会话管理器接口
type SessionManagerInterface interface {
	IsOnline(playerID string) bool
	CanReconnect(token, playerID string) bool
	GetSession(playerID string) interface{}
	SetOnline(playerID string)
}

// RoomManagerInterface 房间管理器接口
type RoomManagerInterface interface {
	LeaveRoom(client ClientInterface)
	CreateRoom(client ClientInterface) (any, error)
	JoinRoom(client ClientInterface, code string) (any, error)
	SetPlayerReady(client ClientInterface, ready bool) error
	GetRoom(code string) any
	GetRoomList() []any
	GetRoomByPlayerID(playerID string) any
	GetActiveGamesCount() int
}

// MatcherInterface 匹配器接口
type MatcherInterface interface {
	AddToQueue(client ClientInterface)
}

// ClientInterface 客户端接口
type ClientInterface interface {
	GetID() string
	GetName() string
	GetRoom() string
	SetRoom(roomCode string)
	SendMessage(msg *protocol.Message)
	Close()
}

// RoomError 房间错误
type RoomError struct {
	Code    int
	Message string
}

func (e *RoomError) Error() string {
	return e.Message
}

// RoomState 房间状态
type RoomState int

const (
	RoomStateWaiting RoomState = iota
	RoomStateReady
	RoomStateBidding
	RoomStatePlaying
	RoomStateEnded
)

// RoomInterface 房间接口 - GameSession 依赖的 Room 方法
type RoomInterface interface {
	// 广播消息
	Broadcast(msg *protocol.Message)

	// 玩家访问
	GetPlayer(id string) RoomPlayerInterface
	GetPlayerOrder() []string
	SetPlayerLandlord(id string)

	// 房间信息
	GetCode() string
	SetState(RoomState)

	// 服务访问
	GetServer() ServerContext
}

// RoomPlayerInterface 房间玩家接口
type RoomPlayerInterface interface {
	GetClient() ClientInterface
}

// ChatLimiterInterface 聊天限流器接口
type ChatLimiterInterface interface {
	AllowChat(playerID string) (bool, string)
}

// Config access
type ConfigInterface any

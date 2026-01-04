package room

import (
	"sync"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/types"
)

const (
	roomCodeLength = 6            // 房间号长度
	roomCodeChars  = "0123456789" // 房间号字符集
)

// RoomPlayer 房间中的玩家
type RoomPlayer struct {
	Client     types.ClientInterface
	Seat       int  // 座位号 0-2
	Ready      bool // 是否准备
	IsLandlord bool // 是否是地主
}

// Room 游戏房间
type Room struct {
	Code        string                 // 房间号
	State       RoomState              // 房间状态
	Players     map[string]*RoomPlayer // 玩家列表
	PlayerOrder []string               // 玩家顺序（按座位）
	CreatedAt   time.Time              // 创建时间

	mu sync.RWMutex
}

// RoomManager 房间管理器
type RoomManager struct {
	redisStore  *storage.RedisStore
	roomTimeout time.Duration
	rooms       map[string]*Room
	mu          sync.RWMutex
}

// NewRoomManager 创建房间管理器
func NewRoomManager(rs *storage.RedisStore, roomTimeout time.Duration) *RoomManager {
	rm := &RoomManager{
		redisStore:  rs,
		roomTimeout: roomTimeout,
		rooms:       make(map[string]*Room),
	}

	// 启动房间清理协程
	go rm.cleanupLoop()

	return rm
}

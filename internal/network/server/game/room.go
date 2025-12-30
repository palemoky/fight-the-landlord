package game

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game/session"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

const (
	// 房间号长度
	roomCodeLength = 6
	// 房间号字符集
	roomCodeChars = "0123456789"
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
	State       types.RoomState        // 房间状态
	Players     map[string]*RoomPlayer // 玩家列表
	PlayerOrder []string               // 玩家顺序（按座位）
	CreatedAt   time.Time              // 创建时间

	game   *session.GameSession // 游戏会话
	server types.ServerContext
	mu     sync.RWMutex
}

// RoomManager 房间管理器
type RoomManager struct {
	server types.ServerContext
	rooms  map[string]*Room
	mu     sync.RWMutex
}

// NewRoomManager 创建房间管理器
func NewRoomManager(s types.ServerContext) *RoomManager {
	rm := &RoomManager{
		server: s,
		rooms:  make(map[string]*Room),
	}

	// 启动房间清理协程
	go rm.cleanupLoop()

	return rm
}

// CreateRoom 创建房间
func (rm *RoomManager) CreateRoom(client types.ClientInterface) (interface{}, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 生成唯一房间号
	code := rm.generateRoomCode()

	room := &Room{
		Code:        code,
		State:       types.RoomStateWaiting,
		Players:     make(map[string]*RoomPlayer),
		PlayerOrder: make([]string, 0, 3),
		CreatedAt:   time.Now(),
		server:      rm.server,
	}

	// 添加创建者
	player := &RoomPlayer{
		Client: client,
		Seat:   0,
		Ready:  false,
	}
	room.Players[client.GetID()] = player
	room.PlayerOrder = append(room.PlayerOrder, client.GetID())
	client.SetRoom(code)

	rm.rooms[code] = room

	// 保存到 Redis
	go func() { _ = rm.server.GetRedisStore().SaveRoom(context.Background(), room) }()

	log.Printf("🏠 房间 %s 已创建，玩家 %s", code, client.GetName())

	return room, nil
}

// JoinRoom 加入房间
func (rm *RoomManager) JoinRoom(client types.ClientInterface, code string) (interface{}, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[code]
	if !exists {
		return nil, ErrRoomNotFound
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.Players) >= 3 {
		return nil, ErrRoomFull
	}

	if room.State != types.RoomStateWaiting {
		return nil, ErrGameStarted
	}

	// 分配座位
	seat := len(room.Players)
	player := &RoomPlayer{
		Client: client,
		Seat:   seat,
		Ready:  false,
	}
	room.Players[client.GetID()] = player
	room.PlayerOrder = append(room.PlayerOrder, client.GetID())
	client.SetRoom(code)

	log.Printf("👤 玩家 %s 加入房间 %s", client.GetName(), code)

	// 通知房间内其他玩家
	room.broadcastExcept(client.GetID(), encoding.MustNewMessage(protocol.MsgPlayerJoined, protocol.PlayerJoinedPayload{
		Player: room.GetPlayerInfo(client.GetID()),
	}))

	// 保存到 Redis
	go func() { _ = rm.server.GetRedisStore().SaveRoom(context.Background(), room) }()

	return room, nil
}

// LeaveRoom 离开房间
func (rm *RoomManager) LeaveRoom(client types.ClientInterface) {
	roomCode := client.GetRoom()
	if roomCode == "" {
		return
	}

	rm.mu.Lock()
	room, exists := rm.rooms[roomCode]
	if !exists {
		rm.mu.Unlock()
		return
	}
	rm.mu.Unlock()

	room.mu.Lock()
	defer room.mu.Unlock()

	player, exists := room.Players[client.GetID()]
	if !exists {
		return
	}

	// 通知其他玩家
	room.broadcastExcept(client.GetID(), encoding.MustNewMessage(protocol.MsgPlayerLeft, protocol.PlayerLeftPayload{
		PlayerID:   client.GetID(),
		PlayerName: client.GetName(),
	}))

	// 移除玩家
	delete(room.Players, client.GetID())
	// 从顺序列表中移除
	for i, id := range room.PlayerOrder {
		if id == client.GetID() {
			room.PlayerOrder = append(room.PlayerOrder[:i], room.PlayerOrder[i+1:]...)
			break
		}
	}
	client.SetRoom("")

	log.Printf("👋 玩家 %s 离开房间 %s (座位 %d)", client.GetName(), roomCode, player.Seat)

	// 如果房间空了，删除房间
	if len(room.Players) == 0 {
		rm.mu.Lock()
		delete(rm.rooms, roomCode)
		rm.mu.Unlock()
		// 从 Redis 删除
		go func() { _ = rm.server.GetRedisStore().DeleteRoom(context.Background(), roomCode) }()
		log.Printf("🏠 房间 %s 已解散", roomCode)
	} else {
		// 更新 Redis
		go func() { _ = rm.server.GetRedisStore().SaveRoom(context.Background(), room) }()
	}
}

// SetPlayerReady 设置玩家准备状态
func (rm *RoomManager) SetPlayerReady(client types.ClientInterface, ready bool) error {
	roomCode := client.GetRoom()
	if roomCode == "" {
		return ErrNotInRoom
	}

	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return ErrRoomNotFound
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	player, exists := room.Players[client.GetID()]
	if !exists {
		return ErrNotInRoom
	}

	player.Ready = ready

	// 广播准备状态
	room.broadcast(encoding.MustNewMessage(protocol.MsgPlayerReady, protocol.PlayerReadyPayload{
		PlayerID: client.GetID(),
		Ready:    ready,
	}))

	// 检查是否所有人都准备好了
	if room.checkAllReady() {
		go room.startGame()
	}

	return nil
}

// GetRoom 获取房间
func (rm *RoomManager) GetRoom(code string) interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

// GetRoomList 获取可加入的房间列表
func (rm *RoomManager) GetRoomList() []interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var rooms []interface{}
	for code, room := range rm.rooms {
		room.mu.RLock()
		// 只返回等待中且未满的房间
		if room.State == types.RoomStateWaiting && len(room.Players) < 3 {
			rooms = append(rooms, protocol.RoomListItem{
				RoomCode:    code,
				PlayerCount: len(room.Players),
				MaxPlayers:  3,
			})
		}
		room.mu.RUnlock()
	}
	return rooms
}

// NotifyPlayerOffline 通知房间内其他玩家某个玩家掉线
func (rm *RoomManager) NotifyPlayerOffline(client types.ClientInterface) {
	roomCode := client.GetRoom()
	if roomCode == "" {
		return
	}

	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return
	}

	room.mu.Lock()

	// 通知其他在线玩家
	for id, player := range room.Players {
		if id != client.GetID() && player.Client != nil {
			player.Client.SendMessage(encoding.MustNewMessage(protocol.MsgPlayerOffline, protocol.PlayerOfflinePayload{
				PlayerID:   client.GetID(),
				PlayerName: client.GetName(),
				Timeout:    20, // 20秒离线等待
			}))
		}
	}

	// 如果游戏进行中，通知 GameSession 暂停该玩家的计时器
	game := room.game
	room.mu.Unlock()

	if game != nil {
		game.PlayerOffline(client.GetID())
	}

	log.Printf("📴 玩家 %s 在房间 %s 中掉线", client.GetName(), roomCode)
}

// ReconnectPlayer 玩家重连到房间
func (rm *RoomManager) ReconnectPlayer(oldClient types.ClientInterface, newClient types.ClientInterface) error {
	roomCode := oldClient.GetRoom()
	if roomCode == "" {
		return nil // 不在房间中，无需重连
	}

	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return ErrRoomNotFound
	}

	room.mu.Lock()

	player, exists := room.Players[oldClient.GetID()]
	if !exists {
		room.mu.Unlock()
		return ErrNotInRoom
	}

	// 更新客户端引用
	player.Client = newClient
	newClient.SetRoom(roomCode)

	// 通知其他玩家该玩家已上线
	for id, p := range room.Players {
		if id != newClient.GetID() && p.Client != nil {
			p.Client.SendMessage(encoding.MustNewMessage(protocol.MsgPlayerOnline, protocol.PlayerOnlinePayload{
				PlayerID:   newClient.GetID(),
				PlayerName: newClient.GetName(),
			}))
		}
	}

	// 如果游戏进行中，通知 GameSession 恢复该玩家的计时器
	game := room.game
	room.mu.Unlock()

	if game != nil {
		game.PlayerOnline(newClient.GetID())
	}

	log.Printf("📶 玩家 %s 重连到房间 %s", newClient.GetName(), roomCode)

	return nil
}

// GetRoomByPlayerID 通过玩家 ID 获取房间
func (rm *RoomManager) GetRoomByPlayerID(playerID string) interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, room := range rm.rooms {
		room.mu.RLock()
		_, exists := room.Players[playerID]
		room.mu.RUnlock()
		if exists {
			return room
		}
	}
	return nil
}

// generateRoomCode 生成房间号
func (rm *RoomManager) generateRoomCode() string {
	for {
		code := make([]byte, roomCodeLength)
		for i := range code {
			code[i] = roomCodeChars[rand.Intn(len(roomCodeChars))]
		}
		codeStr := string(code)
		if _, exists := rm.rooms[codeStr]; !exists {
			return codeStr
		}
	}
}

// cleanupLoop 定期清理超时房间
func (rm *RoomManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rm.cleanup()
	}
}

// cleanup 清理超时房间
func (rm *RoomManager) cleanup() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	timeout := 2 * time.Hour
	now := time.Now()

	for code, room := range rm.rooms {
		room.mu.RLock()
		// 只清理等待状态且超时的房间
		if room.State == types.RoomStateWaiting && now.Sub(room.CreatedAt) > timeout {
			room.mu.RUnlock()
			// 通知所有玩家房间已关闭
			room.broadcast(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, "房间超时已关闭"))
			// 清理玩家状态
			for _, p := range room.Players {
				p.Client.SetRoom("")
			}
			delete(rm.rooms, code)
			log.Printf("🏠 房间 %s 超时已清理", code)
		} else {
			room.mu.RUnlock()
		}
	}
}

// GetActiveGamesCount 获取进行中的游戏数量
func (rm *RoomManager) GetActiveGamesCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	count := 0
	for _, room := range rm.rooms {
		room.mu.RLock()
		// 统计正在游戏中的房间（叫地主、出牌、游戏结束等待清理）
		switch room.State {
		case types.RoomStateBidding, types.RoomStatePlaying, types.RoomStateEnded:
			count++
		}
		room.mu.RUnlock()
	}
	return count
}

// Interface implementations for types.RoomInterface
func (r *Room) GetServer() types.ServerContext { return r.server }

// GetPlayer implements session.RoomInterface
func (r *Room) GetPlayer(id string) session.RoomPlayerInterface {
	return r.Players[id]
}

// GetPlayerOrder implements session.RoomInterface
func (r *Room) GetPlayerOrder() []string {
	return r.PlayerOrder
}

// SetPlayerLandlord implements session.RoomInterface
func (r *Room) SetPlayerLandlord(id string) {
	if player, exists := r.Players[id]; exists {
		player.IsLandlord = true
	}
}

// GetCode implements session.RoomInterface
func (r *Room) GetCode() string {
	return r.Code
}

// SetState implements types.RoomInterface
func (r *Room) SetState(state types.RoomState) {
	r.State = state
}

// SerializeForRedis 为Redis序列化准备数据（提供只读访问）
func (r *Room) SerializeForRedis(serialize func()) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	serialize()
}

// GetGameForSerialization 获取game用于序列化（只读）
func (r *Room) GetGameForSerialization() *session.GameSession {
	return r.game
}

// SetGameSession 设置游戏会话（主要用于测试或状态恢复）
func (r *Room) SetGameSession(gs *session.GameSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.game = gs
}

// RoomPlayer implements session.RoomPlayerInterface
func (rp *RoomPlayer) GetClient() types.ClientInterface {
	return rp.Client
}

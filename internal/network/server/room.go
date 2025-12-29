package server

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

const (
	// 房间号长度
	roomCodeLength = 6
	// 房间号字符集
	roomCodeChars = "0123456789"
)

// RoomState 房间状态
type RoomState int

const (
	RoomStateWaiting RoomState = iota // 等待玩家
	RoomStateReady                    // 准备就绪
	RoomStateBidding                  // 叫地主中
	RoomStatePlaying                  // 游戏中
	RoomStateEnded                    // 游戏结束
)

// RoomPlayer 房间中的玩家
type RoomPlayer struct {
	Client     *Client
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

	game   *GameSession // 游戏会话
	server *Server
	mu     sync.RWMutex
}

// RoomManager 房间管理器
type RoomManager struct {
	server *Server
	rooms  map[string]*Room
	mu     sync.RWMutex
}

// NewRoomManager 创建房间管理器
func NewRoomManager(s *Server) *RoomManager {
	rm := &RoomManager{
		server: s,
		rooms:  make(map[string]*Room),
	}

	// 启动房间清理协程
	go rm.cleanupLoop()

	return rm
}

// CreateRoom 创建房间
func (rm *RoomManager) CreateRoom(client *Client) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 生成唯一房间号
	code := rm.generateRoomCode()

	room := &Room{
		Code:        code,
		State:       RoomStateWaiting,
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
	room.Players[client.ID] = player
	room.PlayerOrder = append(room.PlayerOrder, client.ID)
	client.SetRoom(code)

	rm.rooms[code] = room

	// 保存到 Redis
	go func() { _ = rm.server.redisStore.SaveRoom(context.Background(), room) }()

	log.Printf("🏠 房间 %s 已创建，玩家 %s", code, client.Name)

	return room, nil
}

// JoinRoom 加入房间
func (rm *RoomManager) JoinRoom(client *Client, code string) (*Room, error) {
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

	if room.State != RoomStateWaiting {
		return nil, ErrGameStarted
	}

	// 分配座位
	seat := len(room.Players)
	player := &RoomPlayer{
		Client: client,
		Seat:   seat,
		Ready:  false,
	}
	room.Players[client.ID] = player
	room.PlayerOrder = append(room.PlayerOrder, client.ID)
	client.SetRoom(code)

	log.Printf("👤 玩家 %s 加入房间 %s", client.Name, code)

	// 通知房间内其他玩家
	room.broadcastExcept(client.ID, encoding.MustNewMessage(protocol.MsgPlayerJoined, protocol.PlayerJoinedPayload{
		Player: room.getPlayerInfo(client.ID),
	}))

	// 保存到 Redis
	go func() { _ = rm.server.redisStore.SaveRoom(context.Background(), room) }()

	return room, nil
}

// LeaveRoom 离开房间
func (rm *RoomManager) LeaveRoom(client *Client) {
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

	player, exists := room.Players[client.ID]
	if !exists {
		return
	}

	// 通知其他玩家
	room.broadcastExcept(client.ID, encoding.MustNewMessage(protocol.MsgPlayerLeft, protocol.PlayerLeftPayload{
		PlayerID:   client.ID,
		PlayerName: client.Name,
	}))

	// 移除玩家
	delete(room.Players, client.ID)
	// 从顺序列表中移除
	for i, id := range room.PlayerOrder {
		if id == client.ID {
			room.PlayerOrder = append(room.PlayerOrder[:i], room.PlayerOrder[i+1:]...)
			break
		}
	}
	client.SetRoom("")

	log.Printf("👋 玩家 %s 离开房间 %s (座位 %d)", client.Name, roomCode, player.Seat)

	// 如果房间空了，删除房间
	if len(room.Players) == 0 {
		rm.mu.Lock()
		delete(rm.rooms, roomCode)
		rm.mu.Unlock()
		// 从 Redis 删除
		go func() { _ = rm.server.redisStore.DeleteRoom(context.Background(), roomCode) }()
		log.Printf("🏠 房间 %s 已解散", roomCode)
	} else {
		// 更新 Redis
		go func() { _ = rm.server.redisStore.SaveRoom(context.Background(), room) }()
	}
}

// SetPlayerReady 设置玩家准备状态
func (rm *RoomManager) SetPlayerReady(client *Client, ready bool) error {
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

	player, exists := room.Players[client.ID]
	if !exists {
		return ErrNotInRoom
	}

	player.Ready = ready

	// 广播准备状态
	room.broadcast(encoding.MustNewMessage(protocol.MsgPlayerReady, protocol.PlayerReadyPayload{
		PlayerID: client.ID,
		Ready:    ready,
	}))

	// 检查是否所有人都准备好了
	if room.checkAllReady() {
		go room.startGame()
	}

	return nil
}

// GetRoom 获取房间
func (rm *RoomManager) GetRoom(code string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

// GetRoomList 获取可加入的房间列表
func (rm *RoomManager) GetRoomList() []protocol.RoomListItem {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var rooms []protocol.RoomListItem
	for code, room := range rm.rooms {
		room.mu.RLock()
		// 只返回等待中且未满的房间
		if room.State == RoomStateWaiting && len(room.Players) < 3 {
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
func (rm *RoomManager) NotifyPlayerOffline(client *Client) {
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
		if id != client.ID && player.Client != nil {
			player.Client.SendMessage(encoding.MustNewMessage(protocol.MsgPlayerOffline, protocol.PlayerOfflinePayload{
				PlayerID:   client.ID,
				PlayerName: client.Name,
				Timeout:    20, // 20秒离线等待
			}))
		}
	}

	// 如果游戏进行中，通知 GameSession 暂停该玩家的计时器
	game := room.game
	room.mu.Unlock()

	if game != nil {
		game.PlayerOffline(client.ID)
	}

	log.Printf("📴 玩家 %s 在房间 %s 中掉线", client.Name, roomCode)
}

// ReconnectPlayer 玩家重连到房间
func (rm *RoomManager) ReconnectPlayer(oldClient *Client, newClient *Client) error {
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

	player, exists := room.Players[oldClient.ID]
	if !exists {
		room.mu.Unlock()
		return ErrNotInRoom
	}

	// 更新客户端引用
	player.Client = newClient
	newClient.SetRoom(roomCode)

	// 通知其他玩家该玩家已上线
	for id, p := range room.Players {
		if id != newClient.ID && p.Client != nil {
			p.Client.SendMessage(encoding.MustNewMessage(protocol.MsgPlayerOnline, protocol.PlayerOnlinePayload{
				PlayerID:   newClient.ID,
				PlayerName: newClient.Name,
			}))
		}
	}

	// 如果游戏进行中，通知 GameSession 恢复该玩家的计时器
	game := room.game
	room.mu.Unlock()

	if game != nil {
		game.PlayerOnline(newClient.ID)
	}

	log.Printf("📶 玩家 %s 重连到房间 %s", newClient.Name, roomCode)

	return nil
}

// GetRoomByPlayerID 通过玩家 ID 获取房间
func (rm *RoomManager) GetRoomByPlayerID(playerID string) *Room {
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

	timeout := rm.server.config.Game.RoomTimeoutDuration()
	now := time.Now()

	for code, room := range rm.rooms {
		room.mu.RLock()
		// 只清理等待状态且超时的房间
		if room.State == RoomStateWaiting && now.Sub(room.CreatedAt) > timeout {
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
		case RoomStateBidding, RoomStatePlaying, RoomStateEnded:
			count++
		}
		room.mu.RUnlock()
	}
	return count
}

// --- Room 方法 ---

// Broadcast 广播消息给房间内所有玩家
func (r *Room) Broadcast(msg *protocol.Message) {
	for _, player := range r.Players {
		player.Client.SendMessage(msg)
	}
}

// broadcast 内部使用的广播方法（保留以兼容现有代码）
func (r *Room) broadcast(msg *protocol.Message) {
	r.Broadcast(msg)
}

// broadcastExcept 广播消息给除指定玩家外的所有玩家
func (r *Room) broadcastExcept(excludeID string, msg *protocol.Message) {
	for id, player := range r.Players {
		if id != excludeID {
			player.Client.SendMessage(msg)
		}
	}
}

// checkAllReady 检查是否所有玩家都准备好
func (r *Room) checkAllReady() bool {
	if len(r.Players) < 3 {
		return false
	}
	for _, player := range r.Players {
		if !player.Ready {
			return false
		}
	}
	return true
}

// getPlayerInfo 获取玩家信息
func (r *Room) getPlayerInfo(playerID string) protocol.PlayerInfo {
	player := r.Players[playerID]
	cardsCount := 0
	if r.game != nil {
		cardsCount = r.game.GetPlayerCardsCount(playerID)
	}
	return protocol.PlayerInfo{
		ID:         player.Client.ID,
		Name:       player.Client.Name,
		Seat:       player.Seat,
		Ready:      player.Ready,
		IsLandlord: player.IsLandlord,
		CardsCount: cardsCount,
	}
}

// getAllPlayersInfo 获取所有玩家信息
func (r *Room) getAllPlayersInfo() []protocol.PlayerInfo {
	infos := make([]protocol.PlayerInfo, 0, len(r.Players))
	for _, id := range r.PlayerOrder {
		infos = append(infos, r.getPlayerInfo(id))
	}
	return infos
}

// startGame 开始游戏
func (r *Room) startGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != RoomStateWaiting || len(r.Players) < 3 {
		return
	}

	r.State = RoomStateReady

	// 广播游戏开始
	r.broadcast(encoding.MustNewMessage(protocol.MsgGameStart, protocol.GameStartPayload{
		Players: r.getAllPlayersInfo(),
	}))

	// 创建游戏会话
	r.game = NewGameSession(r)

	// 开始游戏流程
	r.game.Start()

	// 保存到 Redis
	go func() { _ = r.server.redisStore.SaveRoom(context.Background(), r) }()
}

// GetGameSession 获取游戏会话
func (r *Room) GetGameSession() *GameSession {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.game
}

// SaveToRedis 保存房间状态到 Redis
// SaveToRedis 保存房间状态到 Redis
func (r *Room) SaveToRedis(ctx context.Context) error {
	if r.server != nil && r.server.redisStore != nil {
		return r.server.redisStore.SaveRoom(ctx, r)
	}
	return nil
}

// --- 错误定义 ---

type RoomError struct {
	Code    int
	Message string
}

func (e *RoomError) Error() string {
	return e.Message
}

var (
	ErrRoomNotFound = &RoomError{Code: protocol.ErrCodeRoomNotFound, Message: "房间不存在"}
	ErrRoomFull     = &RoomError{Code: protocol.ErrCodeRoomFull, Message: "房间已满"}
	ErrNotInRoom    = &RoomError{Code: protocol.ErrCodeNotInRoom, Message: "您不在房间中"}
	ErrGameStarted  = &RoomError{Code: protocol.ErrCodeGameNotStart, Message: "游戏已开始"}
)

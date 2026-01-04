package room

import (
	"context"
	"log"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/apperrors"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/types"
)

// CreateRoom 创建房间
func (rm *RoomManager) CreateRoom(client types.ClientInterface) (*Room, error) {
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
	go func() { _ = rm.redisStore.SaveRoom(context.Background(), room.Code, room.ToRoomData()) }()

	log.Printf("🏠 房间 %s 已创建，玩家 %s", code, client.GetName())

	return room, nil
}

// JoinRoom 加入房间
func (rm *RoomManager) JoinRoom(client types.ClientInterface, code string) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[code]
	if !exists {
		return nil, apperrors.ErrRoomNotFound
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.Players) >= 3 {
		return nil, apperrors.ErrRoomFull
	}

	if room.State != RoomStateWaiting {
		return nil, apperrors.ErrGameStarted
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
	room.BroadcastExcept(client.GetID(), codec.MustNewMessage(protocol.MsgPlayerJoined, protocol.PlayerJoinedPayload{
		Player: room.GetPlayerInfo(client.GetID()),
	}))

	// 保存到 Redis
	go func() { _ = rm.redisStore.SaveRoom(context.Background(), room.Code, room.ToRoomData()) }()

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
	room.BroadcastExcept(client.GetID(), codec.MustNewMessage(protocol.MsgPlayerLeft, protocol.PlayerLeftPayload{
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
		go func() { _ = rm.redisStore.DeleteRoom(context.Background(), roomCode) }()
		log.Printf("🏠 房间 %s 已解散", roomCode)
	} else {
		// 更新 Redis
		go func() { _ = rm.redisStore.SaveRoom(context.Background(), room.Code, room.ToRoomData()) }()
	}
}

// SetPlayerReady 设置玩家准备状态
func (rm *RoomManager) SetPlayerReady(client types.ClientInterface, ready bool) error {
	roomCode := client.GetRoom()
	if roomCode == "" {
		return apperrors.ErrNotInRoom
	}

	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return apperrors.ErrRoomNotFound
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	player, exists := room.Players[client.GetID()]
	if !exists {
		return apperrors.ErrNotInRoom
	}

	player.Ready = ready

	// 广播准备状态
	room.Broadcast(codec.MustNewMessage(protocol.MsgPlayerReady, protocol.PlayerReadyPayload{
		PlayerID: client.GetID(),
		Ready:    ready,
	}))

	// 检查是否所有人都准备好了
	if room.checkAllReady() {
		if err := room.StartGame(); err == nil {
			go func() { _ = rm.redisStore.SaveRoom(context.Background(), room.Code, room.ToRoomData()) }()
		}
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

// GetActiveGamesCount 获取进行中的游戏数量
func (rm *RoomManager) GetActiveGamesCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	count := 0
	for _, room := range rm.rooms {
		room.mu.RLock()
		// 只统计正在游戏中的房间（叫地主、出牌）
		// RoomStateEnded 不计入，因为游戏已结束只是等待清理
		switch room.State {
		case RoomStateBidding, RoomStatePlaying:
			count++
		}
		room.mu.RUnlock()
	}
	return count
}

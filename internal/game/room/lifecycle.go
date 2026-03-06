package room

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/apperrors"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/types"
)

// SetAllPlayersReady 设置所有玩家准备状态
func (r *Room) SetAllPlayersReady() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, player := range r.Players {
		player.Ready = true
	}
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

	// 标记当前玩家为离线
	if player, exists := room.Players[client.GetID()]; exists {
		player.Client = nil
	}

	// 检查所有玩家是否都离线
	allOffline := true
	for _, player := range room.Players {
		if player.Client != nil {
			allOffline = false
			// 通知其他在线玩家
			player.Client.SendMessage(codec.MustNewMessage(protocol.MsgPlayerOffline, protocol.PlayerOfflinePayload{
				PlayerID:   client.GetID(),
				PlayerName: client.GetName(),
				Timeout:    rm.gameConfig.OfflineWaitTimeout, // 从配置读取
			}))
		}
	}

	// 如果所有玩家都离线，清理房间
	if allOffline {
		log.Printf("🧹 房间 %s 所有玩家已断开连接，清理房间", roomCode)
		room.State = RoomStateEnded
		room.mu.Unlock()

		// 删除房间
		rm.mu.Lock()
		delete(rm.rooms, roomCode)
		rm.mu.Unlock()
		return
	}

	// 如果游戏进行中，通知 GameSession 暂停该玩家的计时器（由外部调用者处理）
	room.mu.Unlock()

	log.Printf("📴 玩家 %s 在房间 %s 中掉线", client.GetName(), roomCode)
}

// ReconnectPlayer 玩家重连到房间
func (rm *RoomManager) ReconnectPlayer(oldClient, newClient types.ClientInterface) error {
	roomCode := oldClient.GetRoom()
	if roomCode == "" {
		return nil // 不在房间中，无需重连
	}

	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return apperrors.ErrRoomNotFound
	}

	room.mu.Lock()

	player, exists := room.Players[oldClient.GetID()]
	if !exists {
		room.mu.Unlock()
		return apperrors.ErrNotInRoom
	}

	// 更新客户端引用
	player.Client = newClient
	newClient.SetRoom(roomCode)

	// 通知其他玩家该玩家已上线
	for id, p := range room.Players {
		if id != newClient.GetID() && p.Client != nil {
			p.Client.SendMessage(codec.MustNewMessage(protocol.MsgPlayerOnline, protocol.PlayerOnlinePayload{
				PlayerID:   newClient.GetID(),
				PlayerName: newClient.GetName(),
			}))
		}
	}

	// 如果游戏进行中，通知 GameSession 恢复该玩家的计时器（由外部调用者处理）
	room.mu.Unlock()

	log.Printf("📶 玩家 %s 重连到房间 %s", newClient.GetName(), roomCode)

	return nil
}

// generateRoomCode 生成房间号
func (rm *RoomManager) generateRoomCode() string {
	for {
		code := make([]byte, roomCodeLength)
		for i := range code {
			code[i] = roomCodeChars[rand.IntN(len(roomCodeChars))]
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

	now := time.Now()

	for code, room := range rm.rooms {
		room.mu.RLock()
		// 只清理等待状态且超时的房间
		if room.State == RoomStateWaiting && now.Sub(room.CreatedAt) > rm.roomTimeout {
			room.mu.RUnlock()
			// 通知所有玩家房间已关闭
			room.Broadcast(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, "房间超时已关闭"))
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

// SerializeForRedis 为Redis序列化准备数据（提供只读访问）
func (r *Room) SerializeForRedis(serialize func()) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	serialize()
}

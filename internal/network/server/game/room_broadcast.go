package game

import (
	"context"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game/session"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

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

// GetPlayerInfo 获取玩家信息
func (r *Room) GetPlayerInfo(playerID string) protocol.PlayerInfo {
	player := r.Players[playerID]
	cardsCount := 0
	if r.game != nil {
		cardsCount = r.game.GetPlayerCardsCount(playerID)
	}
	return protocol.PlayerInfo{
		ID:         player.Client.GetID(),
		Name:       player.Client.GetName(),
		Seat:       player.Seat,
		Ready:      player.Ready,
		IsLandlord: player.IsLandlord,
		CardsCount: cardsCount,
	}
}

// GetAllPlayersInfo 获取所有玩家信息
func (r *Room) GetAllPlayersInfo() []protocol.PlayerInfo {
	infos := make([]protocol.PlayerInfo, 0, len(r.Players))
	for _, id := range r.PlayerOrder {
		infos = append(infos, r.GetPlayerInfo(id))
	}
	return infos
}

// startGame 开始游戏
func (r *Room) startGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != types.RoomStateWaiting || len(r.Players) < 3 {
		return
	}

	r.State = types.RoomStateReady

	// 广播游戏开始
	r.broadcast(encoding.MustNewMessage(protocol.MsgGameStart, protocol.GameStartPayload{
		Players: r.GetAllPlayersInfo(),
	}))

	// 创建游戏会话
	r.game = session.NewGameSession(r)

	// 开始游戏流程
	r.game.Start()

	// 保存到 Redis
	go func() { _ = r.server.GetRedisStore().SaveRoom(context.Background(), r) }()
}

// GetGameSession 获取游戏会话
func (r *Room) GetGameSession() *session.GameSession {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.game
}

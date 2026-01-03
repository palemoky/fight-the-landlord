package handlers

import (
	"log"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game/session"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// handlePing 处理心跳消息
func (h *Handler) handlePing(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.PingPayload](msg)
	if err != nil {
		return
	}

	// 立即回复 pong
	client.SendMessage(codec.MustNewMessage(protocol.MsgPong, protocol.PongPayload{
		ClientTimestamp: payload.Timestamp,
		ServerTimestamp: time.Now().UnixMilli(),
	}))
}

// handleReconnect 处理断线重连
func (h *Handler) handleReconnect(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.ReconnectPayload](msg)
	if err != nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	// 验证重连令牌
	if !h.server.GetSessionManager().CanReconnect(payload.Token, payload.PlayerID) {
		client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, "重连令牌无效或已过期"))
		return
	}

	// 获取旧会话
	sessionInterface := h.server.GetSessionManager().GetSession(payload.PlayerID)
	if sessionInterface == nil {
		client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, "会话不存在"))
		return
	}

	// 类型断言session
	session, ok := sessionInterface.(*session.PlayerSession)
	if !ok {
		client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, "会话类型错误"))
		return
	}

	// 注意：由于ClientInterface不允许修改ID/Name，我们需要通过Server层面处理
	// 这里我们假设client已经是正确的类型，可以进行类型断言
	oldID := client.GetID()

	// 从旧 ID 注销，用新 ID 注册
	h.server.UnregisterClient(oldID)
	h.server.RegisterClient(session.PlayerID, client)

	// 标记会话上线
	h.server.GetSessionManager().SetOnline(session.PlayerID)

	// 构建重连响应
	reconnectPayload := protocol.ReconnectedPayload{
		PlayerID:   session.PlayerID,
		PlayerName: session.PlayerName,
	}

	// 如果在房间中，尝试恢复房间信息
	if session.RoomCode != "" {
		h.tryRestoreRoomState(client, session, &reconnectPayload)
	}

	// 发送重连成功消息
	client.SendMessage(codec.MustNewMessage(protocol.MsgReconnected, reconnectPayload))

	log.Printf("🔄 玩家 %s (%s) 重连成功", session.PlayerName, session.PlayerID)
}

// tryRestoreRoomState 尝试恢复房间状态
func (h *Handler) tryRestoreRoomState(client types.ClientInterface, session *session.PlayerSession, payload *protocol.ReconnectedPayload) {
	roomInterface := h.server.GetRoomManager().GetRoom(session.RoomCode)
	if roomInterface == nil {
		return
	}

	room, ok := roomInterface.(*game.Room)
	if !ok || room == nil {
		return
	}

	oldClient := h.server.GetClientByID(session.PlayerID)
	if oldClient == nil {
		return
	}

	roomMgr, ok := h.server.GetRoomManager().(*game.RoomManager)
	if !ok {
		return
	}

	if err := roomMgr.ReconnectPlayer(oldClient, client); err != nil {
		log.Printf("重连到房间失败: %v", err)
		return
	}

	client.SetRoom(session.RoomCode)
	payload.RoomCode = session.RoomCode

	// 如果游戏正在进行，恢复游戏状态
	if gameSession := room.GetGameSession(); gameSession != nil {
		payload.GameState = gameSession.BuildGameStateDTO(session.PlayerID, h.server.GetSessionManager())
	}
}

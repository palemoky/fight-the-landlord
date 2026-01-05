package handler

import (
	"time"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/types"
)

// handleChat 处理聊天消息
func (h *Handler) handleChat(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.ChatPayload](msg)
	if err != nil {
		return
	}

	// 聊天限流检查
	if h.chatLimiter != nil {
		allowed, reason := h.chatLimiter.AllowChat(client.GetID())
		if !allowed {
			client.SendMessage(codec.NewErrorMessageWithText(
				protocol.ErrCodeRateLimit, reason))
			return
		}
	}

	// 填充发送者信息
	payload.SenderID = client.GetID()
	payload.SenderName = client.GetName()
	payload.Time = time.Now().Unix()

	chatMsg := codec.MustNewMessage(protocol.MsgChat, payload)

	// 大厅聊天：广播给所有大厅玩家
	if payload.Scope != "room" {
		h.server.BroadcastToLobby(chatMsg)
		return
	}

	// 房间聊天：检查房间状态
	roomID := client.GetRoom()
	if roomID == "" {
		client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeNotInRoom, "不在房间中，无法发送房间消息"))
		return
	}

	if h.roomManager == nil {
		return
	}

	room := h.roomManager.GetRoom(roomID)
	if room == nil {
		return
	}

	room.Broadcast(chatMsg)
}

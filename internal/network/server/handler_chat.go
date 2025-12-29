package server

import (
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

// handleChat 处理聊天消息
func (h *Handler) handleChat(client *Client, msg *protocol.Message) {
	payload, err := encoding.ParsePayload[protocol.ChatPayload](msg)
	if err != nil {
		return
	}

	// 聊天限流检查
	allowed, reason := h.server.chatLimiter.AllowChat(client.ID)
	if !allowed {
		client.SendMessage(encoding.NewErrorMessageWithText(
			protocol.ErrCodeRateLimit, reason))
		return
	}

	// 填充发送者信息
	payload.SenderID = client.ID
	payload.SenderName = client.Name
	payload.Time = time.Now().Unix()

	chatMsg := encoding.MustNewMessage(protocol.MsgChat, payload)

	if payload.Scope == "room" {
		// 房间内聊天
		roomID := client.GetRoom()
		if roomID == "" {
			client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeNotInRoom, "不在房间中，无法发送房间消息"))
			return
		}

		room := h.server.roomManager.GetRoom(roomID)
		if room != nil {
			room.Broadcast(chatMsg)
		}
	} else {
		// 大厅聊天 (广播给所有人)
		// 也可以优化为只广播给不在房间的人，或者大厅的人
		// 这里简单处理：广播给所有连接的客户端
		h.server.Broadcast(chatMsg)
	}
}

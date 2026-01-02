package handlers

import (
	"log"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// Handler 消息处理器
type Handler struct {
	server types.ServerContext
}

// NewHandler 创建处理器
func NewHandler(s types.ServerContext) *Handler {
	return &Handler{server: s}
}

// Handle 处理消息
func (h *Handler) Handle(client types.ClientInterface, msg *protocol.Message) {
	switch msg.Type {
	// 连接操作
	case protocol.MsgPing:
		h.handlePing(client, msg)
	case protocol.MsgReconnect:
		h.handleReconnect(client, msg)

	// 房间操作
	case protocol.MsgCreateRoom:
		h.handleCreateRoom(client)
	case protocol.MsgJoinRoom:
		h.handleJoinRoom(client, msg)
	case protocol.MsgLeaveRoom:
		h.handleLeaveRoom(client)
	case protocol.MsgQuickMatch:
		h.handleQuickMatch(client)
	case protocol.MsgReady:
		h.handleReady(client, true)
	case protocol.MsgCancelReady:
		h.handleReady(client, false)

	// 游戏操作
	case protocol.MsgBid:
		h.handleBid(client, msg)
	case protocol.MsgPlayCards:
		h.handlePlayCards(client, msg)
	case protocol.MsgPass:
		h.handlePass(client)

	// 排行榜操作
	case protocol.MsgGetStats:
		h.handleGetStats(client)
	case protocol.MsgGetLeaderboard:
		h.handleGetLeaderboard(client, msg)
	case protocol.MsgGetRoomList:
		h.handleGetRoomList(client)
	case protocol.MsgGetOnlineCount:
		h.handleGetOnlineCount(client)
	case protocol.MsgGetMaintenanceStatus:
		h.handleGetMaintenanceStatus(client)
	case protocol.MsgChat:
		h.handleChat(client, msg)

	default:
		log.Printf("⚠️  未知消息类型: '%s' (来自玩家: %s, ID: %s)", msg.Type, client.GetName(), client.GetID())
		log.Printf("    消息详情: Payload长度=%d bytes", len(msg.Payload))
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeInvalidMsg))
	}
}

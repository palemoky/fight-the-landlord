package handlers

import (
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// handleCreateRoom 处理创建房间
func (h *Handler) handleCreateRoom(client types.ClientInterface) {
	// 维护模式检查
	if h.server.IsMaintenanceMode() {
		client.SendMessage(encoding.NewErrorMessageWithText(
			protocol.ErrCodeServerMaintenance, "服务器维护中，暂停创建房间"))
		return
	}

	// 如果已在房间中，先离开
	if client.GetRoom() != "" {
		h.server.GetRoomManager().LeaveRoom(client)
	}

	roomInterface, err := h.server.GetRoomManager().CreateRoom(client)
	if err != nil {
		client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		return
	}

	room, ok := roomInterface.(*game.Room)
	if !ok || room == nil {
		client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, "创建房间失败"))
		return
	}

	client.SendMessage(encoding.MustNewMessage(protocol.MsgRoomCreated, protocol.RoomCreatedPayload{
		RoomCode: room.Code,
		Player:   room.GetPlayerInfo(client.GetID()),
	}))
}

// handleJoinRoom 处理加入房间
func (h *Handler) handleJoinRoom(client types.ClientInterface, msg *protocol.Message) {
	// 维护模式检查
	if h.server.IsMaintenanceMode() {
		client.SendMessage(encoding.NewErrorMessageWithText(
			protocol.ErrCodeServerMaintenance, "服务器维护中，暂停加入房间"))
		return
	}

	payload, err := encoding.ParsePayload[protocol.JoinRoomPayload](msg)
	if err != nil {
		client.SendMessage(encoding.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	// 如果已在房间中，先离开
	if client.GetRoom() != "" {
		h.server.GetRoomManager().LeaveRoom(client)
	}

	roomInterface, err := h.server.GetRoomManager().JoinRoom(client, payload.RoomCode)
	if err != nil {
		if roomErr, ok := err.(*game.RoomError); ok {
			client.SendMessage(encoding.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
		return
	}

	room, ok := roomInterface.(*game.Room)
	if !ok || room == nil {
		client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, "加入房间失败"))
		return
	}

	client.SendMessage(encoding.MustNewMessage(protocol.MsgRoomJoined, protocol.RoomJoinedPayload{
		RoomCode: room.Code,
		Player:   room.GetPlayerInfo(client.GetID()),
		Players:  room.GetAllPlayersInfo(),
	}))
}

// handleLeaveRoom 处理离开房间
func (h *Handler) handleLeaveRoom(client types.ClientInterface) {
	h.server.GetRoomManager().LeaveRoom(client)
}

// handleQuickMatch 处理快速匹配
func (h *Handler) handleQuickMatch(client types.ClientInterface) {
	// 维护模式检查
	if h.server.IsMaintenanceMode() {
		client.SendMessage(encoding.NewErrorMessageWithText(
			protocol.ErrCodeServerMaintenance, "服务器维护中，暂停快速匹配"))
		return
	}

	// 如果已在房间中，先离开
	if client.GetRoom() != "" {
		h.server.GetRoomManager().LeaveRoom(client)
	}

	h.server.GetMatcher().AddToQueue(client)
}

// handleReady 处理准备
func (h *Handler) handleReady(client types.ClientInterface, ready bool) {
	err := h.server.GetRoomManager().SetPlayerReady(client, ready)
	if err != nil {
		if roomErr, ok := err.(*game.RoomError); ok {
			client.SendMessage(encoding.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

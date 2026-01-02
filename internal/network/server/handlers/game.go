package handlers

import (
	"errors"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// handleBid 处理叫地主
func (h *Handler) handleBid(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.BidPayload](msg)
	if err != nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	roomInterface := h.server.GetRoomManager().GetRoom(client.GetRoom())
	if roomInterface == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	room, ok := roomInterface.(*game.Room)
	if !ok || room == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	gameSession := room.GetGameSession()
	if gameSession == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := gameSession.HandleBid(client.GetID(), payload.Bid); err != nil {
		var roomErr *game.RoomError
		if errors.As(err, &roomErr) {
			client.SendMessage(codec.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// handlePlayCards 处理出牌
func (h *Handler) handlePlayCards(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.PlayCardsPayload](msg)
	if err != nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	roomInterface := h.server.GetRoomManager().GetRoom(client.GetRoom())
	if roomInterface == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	room, ok := roomInterface.(*game.Room)
	if !ok || room == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	gameSession := room.GetGameSession()
	if gameSession == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := gameSession.HandlePlayCards(client.GetID(), payload.Cards); err != nil {
		var roomErr *game.RoomError
		if errors.As(err, &roomErr) {
			client.SendMessage(codec.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// handlePass 处理不出
func (h *Handler) handlePass(client types.ClientInterface) {
	roomInterface := h.server.GetRoomManager().GetRoom(client.GetRoom())
	if roomInterface == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	room, ok := roomInterface.(*game.Room)
	if !ok || room == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	gameSession := room.GetGameSession()
	if gameSession == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := gameSession.HandlePass(client.GetID()); err != nil {
		var roomErr *game.RoomError
		if errors.As(err, &roomErr) {
			client.SendMessage(codec.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

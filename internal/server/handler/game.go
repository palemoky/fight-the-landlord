package handler

import (
	"errors"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/types"

	"github.com/palemoky/fight-the-landlord/internal/apperrors"
)

// handleBid 处理叫地主
func (h *Handler) handleBid(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.BidPayload](msg)
	if err != nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	if h.roomManager == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	room := h.roomManager.GetRoom(client.GetRoom())
	if room == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	gameSession := h.GetGameSession(room.Code)
	if gameSession == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := gameSession.HandleBid(client.GetID(), payload.Bid); err != nil {
		var gameErr *apperrors.GameError
		if errors.As(err, &gameErr) {
			client.SendMessage(codec.NewErrorMessage(gameErr.Code))
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

	if h.roomManager == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	room := h.roomManager.GetRoom(client.GetRoom())
	if room == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	gameSession := h.GetGameSession(room.Code)
	if gameSession == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := gameSession.HandlePlayCards(client.GetID(), payload.Cards); err != nil {
		var gameErr *apperrors.GameError
		if errors.As(err, &gameErr) {
			client.SendMessage(codec.NewErrorMessage(gameErr.Code))
		} else {
			client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// handlePass 处理不出
func (h *Handler) handlePass(client types.ClientInterface) {
	if h.roomManager == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	room := h.roomManager.GetRoom(client.GetRoom())
	if room == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	gameSession := h.GetGameSession(room.Code)
	if gameSession == nil {
		client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := gameSession.HandlePass(client.GetID()); err != nil {
		var gameErr *apperrors.GameError
		if errors.As(err, &gameErr) {
			client.SendMessage(codec.NewErrorMessage(gameErr.Code))
		} else {
			client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

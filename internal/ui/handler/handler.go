// Package handler processes server messages.
package handler

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

// HandleServerMessage dispatches server messages to appropriate handlers.
func HandleServerMessage(m model.Model, msg *protocol.Message) tea.Cmd {
	switch msg.Type {
	// Connection
	case protocol.MsgConnected:
		return handleMsgConnected(m, msg)
	case protocol.MsgReconnected:
		return handleMsgReconnected(m, msg)
	case protocol.MsgPong:
		return handleMsgPong(msg)
	case protocol.MsgError:
		return handleMsgError(m, msg)
	case protocol.MsgOnlineCount:
		return handleMsgOnlineCount(m, msg)

	// Room
	case protocol.MsgRoomCreated:
		return handleMsgRoomCreated(m, msg)
	case protocol.MsgRoomJoined:
		return handleMsgRoomJoined(m, msg)
	case protocol.MsgPlayerJoined:
		return handleMsgPlayerJoined(m, msg)
	case protocol.MsgPlayerLeft:
		return handleMsgPlayerLeft(m, msg)
	case protocol.MsgPlayerReady:
		return handleMsgPlayerReady(m, msg)
	case protocol.MsgPlayerOffline:
		return handleMsgPlayerOffline(m, msg)
	case protocol.MsgPlayerOnline:
		return handleMsgPlayerOnline(m, msg)
	case protocol.MsgRoomListResult:
		return handleMsgRoomListResult(m, msg)

	// Game
	case protocol.MsgGameStart:
		return handleMsgGameStart(m, msg)
	case protocol.MsgDealCards:
		return handleMsgDealCards(m, msg)
	case protocol.MsgBidTurn:
		return handleMsgBidTurn(m, msg)
	case protocol.MsgBidResult:
		return nil
	case protocol.MsgLandlord:
		return handleMsgLandlord(m, msg)
	case protocol.MsgPlayTurn:
		return handleMsgPlayTurn(m, msg)
	case protocol.MsgCardPlayed:
		return handleMsgCardPlayed(m, msg)
	case protocol.MsgPlayerPass:
		return nil
	case protocol.MsgGameOver:
		return handleMsgGameOver(m, msg)

	// Stats
	case protocol.MsgStatsResult:
		return handleMsgStatsResult(m, msg)
	case protocol.MsgLeaderboardResult:
		return handleMsgLeaderboardResult(m, msg)

	// Chat & Maintenance
	case protocol.MsgChat:
		return handleMsgChat(m, msg)
	case protocol.MsgMaintenancePush:
		return handleMsgMaintenancePush(m, msg)
	case protocol.MsgMaintenancePull:
		return handleMsgMaintenancePull(m, msg)
	}

	return nil
}

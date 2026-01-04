// Package handler processes server messages.
package handler

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

// messageHandler 消息处理函数类型
type messageHandler func(m model.Model, msg *protocol.Message) tea.Cmd

// messageHandlers 消息处理器映射表
var messageHandlers = map[protocol.MessageType]messageHandler{
	// Connection
	protocol.MsgConnected:   handleMsgConnected,
	protocol.MsgReconnected: handleMsgReconnected,
	protocol.MsgPong:        func(_ model.Model, msg *protocol.Message) tea.Cmd { return handleMsgPong(msg) },
	protocol.MsgError:       handleMsgError,
	protocol.MsgOnlineCount: handleMsgOnlineCount,

	// Room
	protocol.MsgRoomCreated:    handleMsgRoomCreated,
	protocol.MsgRoomJoined:     handleMsgRoomJoined,
	protocol.MsgPlayerJoined:   handleMsgPlayerJoined,
	protocol.MsgPlayerLeft:     handleMsgPlayerLeft,
	protocol.MsgPlayerReady:    handleMsgPlayerReady,
	protocol.MsgPlayerOffline:  handleMsgPlayerOffline,
	protocol.MsgPlayerOnline:   handleMsgPlayerOnline,
	protocol.MsgRoomListResult: handleMsgRoomListResult,

	// Game
	protocol.MsgGameStart:  handleMsgGameStart,
	protocol.MsgDealCards:  handleMsgDealCards,
	protocol.MsgBidTurn:    handleMsgBidTurn,
	protocol.MsgBidResult:  func(_ model.Model, _ *protocol.Message) tea.Cmd { return nil },
	protocol.MsgLandlord:   handleMsgLandlord,
	protocol.MsgPlayTurn:   handleMsgPlayTurn,
	protocol.MsgCardPlayed: handleMsgCardPlayed,
	protocol.MsgPlayerPass: func(_ model.Model, _ *protocol.Message) tea.Cmd { return nil },
	protocol.MsgGameOver:   handleMsgGameOver,

	// Stats
	protocol.MsgStatsResult:       handleMsgStatsResult,
	protocol.MsgLeaderboardResult: handleMsgLeaderboardResult,

	// Chat & Maintenance
	protocol.MsgChat:            handleMsgChat,
	protocol.MsgMaintenancePush: handleMsgMaintenancePush,
	protocol.MsgMaintenancePull: handleMsgMaintenancePull,
}

// HandleServerMessage dispatches server messages to appropriate handlers.
func HandleServerMessage(m model.Model, msg *protocol.Message) tea.Cmd {
	if handler, ok := messageHandlers[msg.Type]; ok {
		return handler(m, msg)
	}
	return nil
}

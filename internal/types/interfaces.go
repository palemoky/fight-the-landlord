package types

import (
	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

// ServerInterface 服务器上下文接口 - 避免循环依赖
type ServerInterface interface {
	IsMaintenanceMode() bool
	GetOnlineCount() int
	BroadcastToLobby(msg *protocol.Message)
	GetClientByID(id string) ClientInterface
	RegisterClient(id string, client ClientInterface)
	UnregisterClient(id string)
}

// ClientInterface 客户端接口
type ClientInterface interface {
	GetID() string
	GetName() string
	GetRoom() string
	SetRoom(roomCode string)
	SendMessage(msg *protocol.Message)
	Close()
}

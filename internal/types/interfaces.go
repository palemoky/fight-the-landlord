package types

import (
	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

// ServerInterface 定义服务器接口（用于打破循环依赖）
type ServerInterface interface {
	IsMaintenanceMode() bool
	GetOnlineCount() int
	BroadcastToLobby(msg *protocol.Message)
	GetClientByID(id string) ClientInterface
	RegisterClient(id string, client ClientInterface)
	UnregisterClient(id string)
}

// ClientInterface 定义客户端接口
type ClientInterface interface {
	GetID() string
	GetName() string
	GetRoom() string
	SetRoom(code string)
	SendMessage(msg *protocol.Message)
	Close()
}

// ChatLimiter 聊天速率限制器接口
type ChatLimiter interface {
	AllowChat(clientID string) (allowed bool, reason string)
	RemoveClient(clientID string)
}

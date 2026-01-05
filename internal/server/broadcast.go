package server

import "github.com/palemoky/fight-the-landlord/internal/protocol"

// GetOnlineCount 获取在线人数（按需调用）
func (s *Server) GetOnlineCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// Broadcast 广播消息给所有客户端
func (s *Server) Broadcast(msg *protocol.Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		client.SendMessage(msg)
	}
}

// BroadcastToLobby 广播消息给大厅玩家（未在房间内的玩家）
func (s *Server) BroadcastToLobby(msg *protocol.Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.GetRoom() == "" {
			client.SendMessage(msg)
		}
	}
}

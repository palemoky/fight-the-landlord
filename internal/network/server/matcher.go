package server

import (
	"log"
	"sync"
	"time"

	"github.com/palemoky/fight-the-landlord-go/internal/network/protocol"
)

// Matcher 匹配系统
type Matcher struct {
	server *Server
	queue  []*Client
	mu     sync.Mutex
}

// NewMatcher 创建匹配器
func NewMatcher(s *Server) *Matcher {
	return &Matcher{
		server: s,
		queue:  make([]*Client, 0),
	}
}

// AddToQueue 加入匹配队列
func (m *Matcher) AddToQueue(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在队列中
	for _, c := range m.queue {
		if c.ID == client.ID {
			return
		}
	}

	m.queue = append(m.queue, client)
	log.Printf("🔍 玩家 %s 加入匹配队列，当前队列长度: %d", client.Name, len(m.queue))

	// 检查是否可以匹配
	m.tryMatch()
}

// RemoveFromQueue 从匹配队列移除
func (m *Matcher) RemoveFromQueue(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.queue {
		if c.ID == client.ID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			log.Printf("🔍 玩家 %s 离开匹配队列", client.Name)
			return
		}
	}
}

// tryMatch 尝试匹配
func (m *Matcher) tryMatch() {
	if len(m.queue) < 3 {
		return
	}

	// 取出前 3 个玩家
	players := m.queue[:3]
	m.queue = m.queue[3:]

	// 创建房间
	go m.createMatchRoom(players)
}

// createMatchRoom 创建匹配房间
func (m *Matcher) createMatchRoom(players []*Client) {
	// 创建房间（使用第一个玩家）
	room, err := m.server.roomManager.CreateRoom(players[0])
	if err != nil {
		log.Printf("匹配创建房间失败: %v", err)
		// 将玩家放回队列
		m.mu.Lock()
		m.queue = append(players, m.queue...)
		m.mu.Unlock()
		return
	}

	// 其他玩家加入房间
	for _, client := range players[1:] {
		if _, err := m.server.roomManager.JoinRoom(client, room.Code); err != nil {
			log.Printf("匹配加入房间失败: %v", err)
		}
	}

	log.Printf("🎮 匹配成功！房间 %s，玩家: %s, %s, %s",
		room.Code, players[0].Name, players[1].Name, players[2].Name)

	// 给所有玩家发送匹配成功消息和房间信息
	time.Sleep(100 * time.Millisecond) // 短暂延迟确保房间状态同步

	for _, client := range players {
		// 发送加入房间成功消息
		client.SendMessage(protocol.MustNewMessage(protocol.MsgRoomJoined, protocol.RoomJoinedPayload{
			RoomCode: room.Code,
			Player:   room.getPlayerInfo(client.ID),
			Players:  room.getAllPlayersInfo(),
		}))
	}

	// 自动准备所有玩家
	room.mu.Lock()
	for _, player := range room.Players {
		player.Ready = true
	}
	room.mu.Unlock()

	// 广播所有玩家准备状态
	for _, player := range room.Players {
		room.broadcast(protocol.MustNewMessage(protocol.MsgPlayerReady, protocol.PlayerReadyPayload{
			PlayerID: player.Client.ID,
			Ready:    true,
		}))
	}

	// 开始游戏
	go room.startGame()
}

// GetQueueLength 获取队列长度
func (m *Matcher) GetQueueLength() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.queue)
}

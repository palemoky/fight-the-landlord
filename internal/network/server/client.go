package server

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

const (
	// 写入超时
	writeWait = 10 * time.Second

	// 读取超时（pong 等待时间）
	pongWait = 60 * time.Second

	// ping 发送间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 消息最大大小
	maxMessageSize = 4096
)

// Client 代表一个连接的玩家
type Client struct {
	ID     string // 玩家唯一 ID
	Name   string // 玩家昵称
	RoomID string // 当前所在房间 ID
	IP     string // 客户端 IP 地址

	server *Server
	conn   *websocket.Conn
	send   chan []byte

	mu     sync.RWMutex
	closed bool
}

// NewClient 创建新客户端
func NewClient(s *Server, conn *websocket.Conn) *Client {
	return &Client{
		ID:     uuid.New().String(),
		Name:   GenerateNickname(),
		server: s,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

// ReadPump 从 WebSocket 读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.handleDisconnect()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("读取错误: %v", err)
			}
			break
		}

		// 消息速率限制检查
		allowed, warning := c.server.messageLimiter.AllowMessage(c.ID)
		if !allowed {
			log.Printf("⚠️ 客户端 %s (IP: %s) 消息过于频繁", c.Name, c.IP)
			c.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeRateLimit, "消息发送过于频繁"))
			// 如果警告次数过多，断开连接
			if c.server.messageLimiter.GetWarningCount(c.ID) > 5 {
				log.Printf("🚫 客户端 %s 因多次超速被断开连接", c.Name)
				break
			}
			continue
		}
		if warning {
			c.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeRateLimit, "请求过于频繁，请放慢速度"))
		}

		// 解析消息
		msg, err := protocol.Decode(message)
		if err != nil {
			log.Printf("消息解析错误: %v", err)
			c.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeInvalidMsg))
			continue
		}

		// 交给处理器处理
		c.server.handler.Handle(c, msg)
	}
}

// WritePump 向 WebSocket 写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 通道已关闭
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 发送消息给客户端
func (c *Client) SendMessage(msg *protocol.Message) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	data, err := msg.Encode()
	if err != nil {
		log.Printf("消息编码错误: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		// 发送缓冲区已满，关闭连接
		log.Printf("客户端 %s 发送缓冲区已满", c.ID)
		c.Close()
	}
}

// handleDisconnect 处理断开连接
func (c *Client) handleDisconnect() {
	// 标记会话为离线状态
	c.server.sessionManager.SetOffline(c.ID)

	// 如果在房间中，通知房间玩家掉线（但不移除）
	if c.RoomID != "" {
		c.server.roomManager.NotifyPlayerOffline(c)
	}

	// 如果在匹配队列中，移除
	c.server.matcher.RemoveFromQueue(c)

	// 从服务器注销连接（但保留会话）
	c.server.unregisterClient(c)
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.send)
	}
}

// SetRoom 设置客户端所在房间
func (c *Client) SetRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RoomID = roomID
}

// GetRoom 获取客户端所在房间
func (c *Client) GetRoom() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RoomID
}

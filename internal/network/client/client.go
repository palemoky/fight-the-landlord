package client

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/palemoky/fight-the-landlord-go/internal/network/protocol"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	// 心跳检测间隔
	heartbeatInterval = 5 * time.Second
	// 最大重连次数
	maxReconnectAttempts = 5
	// 重连间隔
	reconnectInterval = 2 * time.Second
)

// Client WebSocket 客户端
type Client struct {
	ServerURL string
	conn      *websocket.Conn
	send      chan []byte
	receive   chan *protocol.Message
	done      chan struct{}

	PlayerID       string
	PlayerName     string
	ReconnectToken string // 重连令牌

	// 网络延迟（毫秒）
	Latency int64

	// 回调
	OnMessage       func(*protocol.Message) // 消息回调
	OnError         func(error)             // 错误回调
	OnClose         func()                  // 关闭回调
	OnReconnect     func()                  // 重连成功回调
	OnLatencyUpdate func(int64)             // 延迟更新回调

	mu             sync.RWMutex
	closed         bool
	reconnecting   atomic.Bool
	reconnectCount int
}

// NewClient 创建客户端
func NewClient(serverURL string) *Client {
	return &Client{
		ServerURL: serverURL,
		send:      make(chan []byte, 256),
		receive:   make(chan *protocol.Message, 256),
		done:      make(chan struct{}),
	}
}

// Connect 连接服务器
func (c *Client) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.ServerURL, nil)
	if err != nil {
		return err
	}

	c.conn = conn

	// 启动读写协程
	go c.readPump()
	go c.writePump()

	return nil
}

// readPump 从服务器读取消息
func (c *Client) readPump() {
	defer func() {
		// 尝试重连
		if c.ReconnectToken != "" && !c.reconnecting.Load() {
			go c.tryReconnect()
		} else {
			c.Close()
			if c.OnClose != nil {
				c.OnClose()
			}
		}
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if c.OnError != nil {
					c.OnError(err)
				}
			}
			return
		}

		msg, err := protocol.Decode(message)
		if err != nil {
			log.Printf("消息解析错误: %v", err)
			continue
		}

		// 处理连接成功消息
		if msg.Type == protocol.MsgConnected {
			var payload protocol.ConnectedPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				c.PlayerID = payload.PlayerID
				c.PlayerName = payload.PlayerName
				c.ReconnectToken = payload.ReconnectToken
			}
		}

		// 处理重连成功消息
		if msg.Type == protocol.MsgReconnected {
			c.reconnecting.Store(false)
			c.reconnectCount = 0
			if c.OnReconnect != nil {
				c.OnReconnect()
			}
		}

		// 处理 pong 消息计算延迟
		if msg.Type == protocol.MsgPong {
			var payload protocol.PongPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				latency := time.Now().UnixMilli() - payload.ClientTimestamp
				c.Latency = latency
				if c.OnLatencyUpdate != nil {
					c.OnLatencyUpdate(latency)
				}
			}
		}

		// 回调处理
		if c.OnMessage != nil {
			c.OnMessage(msg)
		}

		// 同时发送到 channel
		select {
		case c.receive <- msg:
		default:
		}
	}
}

// writePump 向服务器写入消息
func (c *Client) writePump() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

// SendMessage 发送消息
func (c *Client) SendMessage(msg *protocol.Message) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return errors.New("connection closed")
	}
	c.mu.RUnlock()

	data, err := msg.Encode()
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

// Receive 接收消息 (阻塞)
func (c *Client) Receive() (*protocol.Message, error) {
	select {
	case msg := <-c.receive:
		return msg, nil
	case <-c.done:
		return nil, errors.New("connection closed")
	}
}

// ReceiveWithTimeout 带超时接收消息
func (c *Client) ReceiveWithTimeout(timeout time.Duration) (*protocol.Message, error) {
	select {
	case msg := <-c.receive:
		return msg, nil
	case <-time.After(timeout):
		return nil, errors.New("receive timeout")
	case <-c.done:
		return nil, errors.New("connection closed")
	}
}

// Close 关闭连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.done)
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

// IsConnected 是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.closed && c.conn != nil
}

// --- 便捷方法 ---

// CreateRoom 创建房间
func (c *Client) CreateRoom() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgCreateRoom, nil))
}

// JoinRoom 加入房间
func (c *Client) JoinRoom(roomCode string) error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgJoinRoom, protocol.JoinRoomPayload{
		RoomCode: roomCode,
	}))
}

// LeaveRoom 离开房间
func (c *Client) LeaveRoom() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgLeaveRoom, nil))
}

// QuickMatch 快速匹配
func (c *Client) QuickMatch() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgQuickMatch, nil))
}

// Ready 准备
func (c *Client) Ready() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgReady, nil))
}

// CancelReady 取消准备
func (c *Client) CancelReady() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgCancelReady, nil))
}

// Bid 叫地主
func (c *Client) Bid(bid bool) error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgBid, protocol.BidPayload{
		Bid: bid,
	}))
}

// PlayCards 出牌
func (c *Client) PlayCards(cards []protocol.CardInfo) error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgPlayCards, protocol.PlayCardsPayload{
		Cards: cards,
	}))
}

// Pass 不出
func (c *Client) Pass() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgPass, nil))
}

// GetStats 获取个人统计
func (c *Client) GetStats() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgGetStats, nil))
}

// GetLeaderboard 获取排行榜
func (c *Client) GetLeaderboard(leaderboardType string, offset, limit int) error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgGetLeaderboard, protocol.GetLeaderboardPayload{
		Type:   leaderboardType,
		Offset: offset,
		Limit:  limit,
	}))
}

// Ping 发送心跳
func (c *Client) Ping() error {
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgPing, protocol.PingPayload{
		Timestamp: time.Now().UnixMilli(),
	}))
}

// Reconnect 手动发送重连请求
func (c *Client) Reconnect() error {
	if c.ReconnectToken == "" || c.PlayerID == "" {
		return errors.New("no reconnect token")
	}
	return c.SendMessage(protocol.MustNewMessage(protocol.MsgReconnect, protocol.ReconnectPayload{
		Token:    c.ReconnectToken,
		PlayerID: c.PlayerID,
	}))
}

// StartHeartbeat 启动心跳检测
func (c *Client) StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if c.IsConnected() {
					c.Ping()
				}
			case <-c.done:
				return
			}
		}
	}()
}

// tryReconnect 尝试重连
func (c *Client) tryReconnect() {
	if c.reconnecting.Load() {
		return
	}
	c.reconnecting.Store(true)

	for c.reconnectCount < maxReconnectAttempts {
		c.reconnectCount++
		log.Printf("🔄 尝试重连 (%d/%d)...", c.reconnectCount, maxReconnectAttempts)

		time.Sleep(reconnectInterval)

		// 创建新连接
		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}

		conn, _, err := dialer.Dial(c.ServerURL, nil)
		if err != nil {
			log.Printf("重连失败: %v", err)
			continue
		}

		// 重置状态
		c.mu.Lock()
		c.conn = conn
		c.closed = false
		c.send = make(chan []byte, 256)
		c.receive = make(chan *protocol.Message, 256)
		c.done = make(chan struct{})
		c.mu.Unlock()

		// 启动读写协程
		go c.readPump()
		go c.writePump()

		// 发送重连请求
		time.Sleep(100 * time.Millisecond)
		if err := c.Reconnect(); err != nil {
			log.Printf("发送重连请求失败: %v", err)
			c.conn.Close()
			continue
		}

		log.Printf("✅ 重连成功")
		return
	}

	// 重连失败
	log.Printf("❌ 重连失败，已达最大尝试次数")
	c.reconnecting.Store(false)
	c.Close()
	if c.OnClose != nil {
		c.OnClose()
	}
}

// GetLatency 获取当前延迟（毫秒）
func (c *Client) GetLatency() int64 {
	return c.Latency
}

// IsReconnecting 是否正在重连
func (c *Client) IsReconnecting() bool {
	return c.reconnecting.Load()
}

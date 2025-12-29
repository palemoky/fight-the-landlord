package client

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
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
	OnMessage       func(*protocol.Message)     // 消息回调
	OnError         func(error)                 // 错误回调
	OnClose         func()                      // 关闭回调
	OnReconnecting  func(attempt, maxTries int) // 正在重连回调
	OnReconnect     func()                      // 重连成功回调
	OnLatencyUpdate func(int64)                 // 延迟更新回调

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
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: false,
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

// SendMessage 发送消息
func (c *Client) SendMessage(msg *protocol.Message) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return errors.New("connection closed")
	}
	c.mu.RUnlock()

	data, err := encoding.Encode(msg)
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
			_ = c.conn.Close()
		}
	}
}

// IsConnected 是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.closed && c.conn != nil
}

package client

import (
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/palemoky/fight-the-landlord/internal/logger"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

// Reconnect 手动发送重连请求
func (c *Client) Reconnect() error {
	if c.ReconnectToken == "" || c.PlayerID == "" {
		return errors.New("no reconnect token")
	}
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgReconnect, protocol.ReconnectPayload{
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
					_ = c.Ping()
				}
			case <-c.done:
				return
			}
		}
	}()
}

// tryReconnect 尝试重连
func (c *Client) tryReconnect() {
	defer func() {
		if r := recover(); r != nil {
			logger.LogPanic(r)
			log.Printf("[PANIC] tryReconnect panic recovered: %v", r)
			c.reconnecting.Store(false)
		}
	}()

	if c.reconnecting.Load() {
		return
	}
	c.reconnecting.Store(true)

	// 指数退避重连策略
	backoff := reconnectInterval

	for c.reconnectCount < maxReconnectAttempts {
		c.reconnectCount++
		// 通过回调通知 UI 正在重连
		if c.OnReconnecting != nil {
			c.OnReconnecting(c.reconnectCount, maxReconnectAttempts)
		}

		time.Sleep(backoff)

		// 计算下一次退避时间 (最大 30 秒)
		backoff *= 2
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}

		// 创建新连接
		dialer := websocket.Dialer{
			HandshakeTimeout:  10 * time.Second,
			EnableCompression: false,
		}

		conn, _, err := dialer.Dial(c.ServerURL, nil)
		if err != nil {
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
			_ = c.conn.Close()
			continue
		}

		// 重连成功（通过 MsgReconnected 消息通知 UI）
		return
	}

	// 重连失败
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

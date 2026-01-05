package transport

import (
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/palemoky/fight-the-landlord/internal/logger"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
)

// readPump 从服务器读取消息
func (c *Client) readPump() {
	defer c.handleReadExit()

	c.setupPongHandler()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.handleReadError(err)
			return
		}

		msg, err := codec.Decode(message)
		if err != nil {
			log.Printf("消息解析错误: %v", err)
			continue
		}

		c.processMessage(msg)
	}
}

func (c *Client) handleReadExit() {
	if r := recover(); r != nil {
		logger.LogPanic(r)
		log.Printf("[PANIC] readPump panic recovered: %v", r)
	}
	// 尝试重连
	if c.ReconnectToken != "" && !c.reconnecting.Load() {
		go c.tryReconnect()
	} else {
		c.Close()
		if c.OnClose != nil {
			c.OnClose()
		}
	}
}

func (c *Client) setupPongHandler() {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
}

func (c *Client) handleReadError(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		if c.OnError != nil {
			c.OnError(err)
		}
	}
}

func (c *Client) processMessage(msg *protocol.Message) {
	isReconnected := c.handleInternalMessage(msg)

	// 回调处理
	if c.OnMessage != nil {
		c.OnMessage(msg)
	}

	// 同时发送到 channel
	select {
	case c.receive <- msg:
	default:
	}

	// 重连成功回调放在最后，确保消息已经发送到 channel
	if isReconnected && c.OnReconnect != nil {
		c.OnReconnect()
	}
}

func (c *Client) handleInternalMessage(msg *protocol.Message) bool {
	switch msg.Type {
	case protocol.MsgConnected:
		var payload protocol.ConnectedPayload
		if err := payloadconv.DecodePayload(msg.Type, msg.Payload, &payload); err == nil {
			c.PlayerID = payload.PlayerID
			c.PlayerName = payload.PlayerName
			c.ReconnectToken = payload.ReconnectToken
		}
	case protocol.MsgReconnected:
		c.reconnecting.Store(false)
		c.reconnectCount = 0
		return true
	case protocol.MsgPong:
		var payload protocol.PongPayload
		if err := payloadconv.DecodePayload(msg.Type, msg.Payload, &payload); err == nil {
			latency := time.Now().UnixMilli() - payload.ClientTimestamp
			c.Latency = latency
			if c.OnLatencyUpdate != nil {
				c.OnLatencyUpdate(latency)
			}
		}
	}
	return false
}

// writePump 向服务器写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil {
			logger.LogPanic(r)
			log.Printf("[PANIC] writePump panic recovered: %v", r)
		}
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

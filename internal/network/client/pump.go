package client

import (
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/palemoky/fight-the-landlord/internal/logger"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

// readPump 从服务器读取消息
func (c *Client) readPump() {
	defer func() {
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
	}()

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
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

		msg, err := encoding.Decode(message)
		if err != nil {
			log.Printf("消息解析错误: %v", err)
			continue
		}

		// 处理连接成功消息
		if msg.Type == protocol.MsgConnected {
			var payload protocol.ConnectedPayload
			if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err == nil {
				c.PlayerID = payload.PlayerID
				c.PlayerName = payload.PlayerName
				c.ReconnectToken = payload.ReconnectToken
			}
		}

		// 处理重连成功消息 - 标记状态但不立即回调
		isReconnected := false
		if msg.Type == protocol.MsgReconnected {
			c.reconnecting.Store(false)
			c.reconnectCount = 0
			isReconnected = true
		}

		// 处理 pong 消息计算延迟
		if msg.Type == protocol.MsgPong {
			var payload protocol.PongPayload
			if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err == nil {
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

		// 重连成功回调放在最后，确保消息已经发送到 channel
		if isReconnected && c.OnReconnect != nil {
			c.OnReconnect()
		}
	}
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

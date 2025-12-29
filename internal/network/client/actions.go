package client

import (
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

// --- 便捷方法 ---

// CreateRoom 创建房间
func (c *Client) CreateRoom() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgCreateRoom, nil))
}

// JoinRoom 加入房间
func (c *Client) JoinRoom(roomCode string) error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgJoinRoom, protocol.JoinRoomPayload{
		RoomCode: roomCode,
	}))
}

// LeaveRoom 离开房间
func (c *Client) LeaveRoom() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgLeaveRoom, nil))
}

// QuickMatch 快速匹配
func (c *Client) QuickMatch() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgQuickMatch, nil))
}

// Ready 准备
func (c *Client) Ready() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgReady, nil))
}

// CancelReady 取消准备
func (c *Client) CancelReady() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgCancelReady, nil))
}

// Bid 叫地主
func (c *Client) Bid(bid bool) error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgBid, protocol.BidPayload{
		Bid: bid,
	}))
}

// PlayCards 出牌
func (c *Client) PlayCards(cards []protocol.CardInfo) error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgPlayCards, protocol.PlayCardsPayload{
		Cards: cards,
	}))
}

// Pass 不出
func (c *Client) Pass() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgPass, nil))
}

// GetStats 获取个人统计
func (c *Client) GetStats() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgGetStats, nil))
}

// GetLeaderboard 获取排行榜
func (c *Client) GetLeaderboard(leaderboardType string, offset, limit int) error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgGetLeaderboard, protocol.GetLeaderboardPayload{
		Type:   leaderboardType,
		Offset: offset,
		Limit:  limit,
	}))
}

// GetRoomList 获取房间列表
func (c *Client) GetRoomList() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgGetRoomList, nil))
}

// Ping 发送心跳
func (c *Client) Ping() error {
	return c.SendMessage(encoding.MustNewMessage(protocol.MsgPing, protocol.PingPayload{
		Timestamp: time.Now().UnixMilli(),
	}))
}

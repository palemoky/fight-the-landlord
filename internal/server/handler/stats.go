package handler

import (
	"context"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/types"
)

// --- 排行榜处理 ---

// handleGetStats 获取个人统计
func (h *Handler) handleGetStats(client types.ClientInterface) {
	ctx := context.Background()
	playerStats, err := h.leaderboard.GetPlayerStats(ctx, client.GetID())
	if err != nil {
		client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, "获取统计失败"))
		return
	}

	if playerStats == nil {
		// 没有统计数据，返回空数据
		client.SendMessage(codec.MustNewMessage(protocol.MsgStatsResult, protocol.StatsResultPayload{
			PlayerID:   client.GetID(),
			PlayerName: client.GetName(),
		}))
		return
	}

	// 获取排名
	rank, _ := h.leaderboard.GetPlayerRank(ctx, client.GetID())

	winRate := 0.0
	if playerStats.TotalGames > 0 {
		winRate = float64(playerStats.Wins) / float64(playerStats.TotalGames) * 100
	}

	client.SendMessage(codec.MustNewMessage(protocol.MsgStatsResult, protocol.StatsResultPayload{
		PlayerID:      playerStats.PlayerID,
		PlayerName:    playerStats.PlayerName,
		TotalGames:    playerStats.TotalGames,
		Wins:          playerStats.Wins,
		Losses:        playerStats.Losses,
		WinRate:       winRate,
		LandlordGames: playerStats.LandlordGames,
		LandlordWins:  playerStats.LandlordWins,
		FarmerGames:   playerStats.FarmerGames,
		FarmerWins:    playerStats.FarmerWins,
		Score:         playerStats.Score,
		Rank:          int(rank),
		CurrentStreak: playerStats.CurrentStreak,
		MaxWinStreak:  playerStats.MaxWinStreak,
	}))
}

// handleGetLeaderboard 获取排行榜
func (h *Handler) handleGetLeaderboard(client types.ClientInterface, msg *protocol.Message) {
	payload, err := codec.ParsePayload[protocol.GetLeaderboardPayload](msg)
	if err != nil {
		// 默认获取总排行榜前 10
		payload = &protocol.GetLeaderboardPayload{
			Type:   "total",
			Offset: 0,
			Limit:  10,
		}
	}

	// 限制请求数量
	if payload.Limit <= 0 || payload.Limit > 50 {
		payload.Limit = 10
	}
	if payload.Offset < 0 {
		payload.Offset = 0
	}

	entries, err := h.leaderboard.GetLeaderboard(context.Background(), payload.Limit)
	if err != nil {
		client.SendMessage(codec.NewErrorMessageWithText(protocol.ErrCodeUnknown, "获取排行榜失败"))
		return
	}

	// 转换为协议格式
	protocolEntries := make([]protocol.LeaderboardEntry, 0, len(entries))
	for _, entry := range entries {
		protocolEntries = append(protocolEntries, protocol.LeaderboardEntry{
			Rank:       entry.Rank,
			PlayerID:   entry.PlayerID,
			PlayerName: entry.PlayerName,
			Score:      entry.Score,
			Wins:       entry.Wins,
			WinRate:    entry.WinRate,
		})
	}

	client.SendMessage(codec.MustNewMessage(protocol.MsgLeaderboardResult, protocol.LeaderboardResultPayload{
		Type:    payload.Type,
		Entries: protocolEntries,
	}))
}

// handleGetRoomList 获取房间列表
func (h *Handler) handleGetRoomList(client types.ClientInterface) {
	rooms := h.roomManager.GetRoomList()

	client.SendMessage(codec.MustNewMessage(protocol.MsgRoomListResult, protocol.RoomListResultPayload{
		Rooms: rooms,
	}))
}

// handleGetOnlineCount 获取在线人数（按需）
func (h *Handler) handleGetOnlineCount(client types.ClientInterface) {
	count := h.server.GetOnlineCount()

	client.SendMessage(codec.MustNewMessage(protocol.MsgOnlineCount, protocol.OnlineCountPayload{
		Count: count,
	}))
}

// handleGetMaintenanceStatus 获取维护状态
func (h *Handler) handleGetMaintenanceStatus(client types.ClientInterface) {
	maintenance := h.server.IsMaintenanceMode()

	client.SendMessage(codec.MustNewMessage(protocol.MsgMaintenancePull, protocol.MaintenanceStatusPayload{
		Maintenance: maintenance,
	}))
}

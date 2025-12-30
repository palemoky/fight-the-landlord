package handlers

import (
	"context"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/network/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// --- 排行榜处理 ---

// handleGetStats 获取个人统计
func (h *Handler) handleGetStats(client types.ClientInterface) {
	ctx := context.Background()
	stats, err := h.server.GetLeaderboard().GetPlayerStats(ctx, client.GetID())
	if err != nil {
		client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, "获取统计失败"))
		return
	}

	if stats == nil {
		// 没有统计数据，返回空数据
		client.SendMessage(encoding.MustNewMessage(protocol.MsgStatsResult, protocol.StatsResultPayload{
			PlayerID:   client.GetID(),
			PlayerName: client.GetName(),
		}))
		return
	}

	// 类型断言
	playerStats, ok := stats.(*storage.PlayerStats)
	if !ok {
		client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, "统计数据格式错误"))
		return
	}

	// 获取排名
	rank, _ := h.server.GetLeaderboard().GetPlayerRank(ctx, client.GetID())

	winRate := 0.0
	if playerStats.TotalGames > 0 {
		winRate = float64(playerStats.Wins) / float64(playerStats.TotalGames) * 100
	}

	client.SendMessage(encoding.MustNewMessage(protocol.MsgStatsResult, protocol.StatsResultPayload{
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
	payload, err := encoding.ParsePayload[protocol.GetLeaderboardPayload](msg)
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

	entries, err := h.server.GetLeaderboard().GetLeaderboard(context.Background(), payload.Limit)
	if err != nil {
		client.SendMessage(encoding.NewErrorMessageWithText(protocol.ErrCodeUnknown, "获取排行榜失败"))
		return
	}

	// 转换为协议格式
	protocolEntries := make([]protocol.LeaderboardEntry, 0, len(entries))
	for _, e := range entries {
		if entry, ok := e.(*storage.LeaderboardEntry); ok {
			protocolEntries = append(protocolEntries, protocol.LeaderboardEntry{
				Rank:       entry.Rank,
				PlayerID:   entry.PlayerID,
				PlayerName: entry.PlayerName,
				Score:      entry.Score,
				Wins:       entry.Wins,
				WinRate:    entry.WinRate,
			})
		}
	}

	client.SendMessage(encoding.MustNewMessage(protocol.MsgLeaderboardResult, protocol.LeaderboardResultPayload{
		Type:    payload.Type,
		Entries: protocolEntries,
	}))
}

// handleGetRoomList 获取房间列表
func (h *Handler) handleGetRoomList(client types.ClientInterface) {
	roomsInterface := h.server.GetRoomManager().GetRoomList()

	// 转换为协议格式
	rooms := make([]protocol.RoomListItem, 0, len(roomsInterface))
	for _, r := range roomsInterface {
		if item, ok := r.(protocol.RoomListItem); ok {
			rooms = append(rooms, item)
		}
	}

	client.SendMessage(encoding.MustNewMessage(protocol.MsgRoomListResult, protocol.RoomListResultPayload{
		Rooms: rooms,
	}))
}

// handleGetOnlineCount 获取在线人数（按需）
func (h *Handler) handleGetOnlineCount(client types.ClientInterface) {
	count := h.server.GetOnlineCount()

	client.SendMessage(encoding.MustNewMessage(protocol.MsgOnlineCount, protocol.OnlineCountPayload{
		Count: count,
	}))
}

// handleGetMaintenanceStatus 获取维护状态
func (h *Handler) handleGetMaintenanceStatus(client types.ClientInterface) {
	maintenance := h.server.IsMaintenanceMode()

	client.SendMessage(encoding.MustNewMessage(protocol.MsgMaintenanceStatus, protocol.MaintenanceStatusPayload{
		Maintenance: maintenance,
	}))
}

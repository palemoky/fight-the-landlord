package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis key
	playerStatsKey    = "player:stats:"
	leaderboardKey    = "leaderboard:score"
	dailyLeaderboard  = "leaderboard:daily:"
	weeklyLeaderboard = "leaderboard:weekly:"
)

// PlayerStats 玩家统计数据
type PlayerStats struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`

	// 总计
	TotalGames int `json:"total_games"` // 总场次
	Wins       int `json:"wins"`        // 胜场
	Losses     int `json:"losses"`      // 败场

	// 地主/农民分开统计
	LandlordGames int `json:"landlord_games"` // 地主场次
	LandlordWins  int `json:"landlord_wins"`  // 地主胜场
	FarmerGames   int `json:"farmer_games"`   // 农民场次
	FarmerWins    int `json:"farmer_wins"`    // 农民胜场

	// 积分
	Score int `json:"score"` // 当前积分

	// 连胜/连败
	CurrentStreak int `json:"current_streak"` // 正数为连胜，负数为连败
	MaxWinStreak  int `json:"max_win_streak"` // 最大连胜

	// 时间
	LastPlayedAt int64 `json:"last_played_at"` // 最后游戏时间
	CreatedAt    int64 `json:"created_at"`     // 首次游戏时间
}

// 积分规则
const (
	WinAsLandlord  = 30  // 地主获胜
	WinAsFarmer    = 15  // 农民获胜
	LoseAsLandlord = -20 // 地主失败
	LoseAsFarmer   = -10 // 农民失败

	// 连胜加成
	StreakBonus3  = 5  // 3 连胜加成
	StreakBonus5  = 10 // 5 连胜加成
	StreakBonus10 = 20 // 10 连胜加成
)

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank       int     `json:"rank"`
	PlayerID   string  `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Score      int     `json:"score"`
	Wins       int     `json:"wins"`
	WinRate    float64 `json:"win_rate"`
}

// LeaderboardManager 排行榜管理器
type LeaderboardManager struct {
	redis *redis.Client
}

// NewLeaderboardManager 创建排行榜管理器
func NewLeaderboardManager(client *redis.Client) *LeaderboardManager {
	return &LeaderboardManager{redis: client}
}

// GetPlayerStats 获取玩家统计
func (lm *LeaderboardManager) GetPlayerStats(ctx context.Context, playerID string) (*PlayerStats, error) {
	key := playerStatsKey + playerID
	data, err := lm.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var stats PlayerStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// SavePlayerStats 保存玩家统计
func (lm *LeaderboardManager) SavePlayerStats(ctx context.Context, stats *PlayerStats) error {
	key := playerStatsKey + stats.PlayerID
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return lm.redis.Set(ctx, key, data, 0).Err()
}

// RecordGameResult 记录游戏结果
func (lm *LeaderboardManager) RecordGameResult(ctx context.Context, playerID, playerName string, isLandlord, isWinner bool) error {
	// 获取或创建玩家统计
	stats, err := lm.GetPlayerStats(ctx, playerID)
	if err != nil {
		return err
	}
	if stats == nil {
		stats = &PlayerStats{
			PlayerID:   playerID,
			PlayerName: playerName,
			CreatedAt:  time.Now().Unix(),
		}
	}

	// 更新名称（可能已更改）
	stats.PlayerName = playerName
	stats.TotalGames++
	stats.LastPlayedAt = time.Now().Unix()

	// 计算积分变化
	var scoreChange int

	if isLandlord {
		stats.LandlordGames++
		if isWinner {
			stats.LandlordWins++
			stats.Wins++
			scoreChange = WinAsLandlord
			stats.CurrentStreak = max(1, stats.CurrentStreak+1)
		} else {
			stats.Losses++
			scoreChange = LoseAsLandlord
			stats.CurrentStreak = min(-1, stats.CurrentStreak-1)
		}
	} else {
		stats.FarmerGames++
		if isWinner {
			stats.FarmerWins++
			stats.Wins++
			scoreChange = WinAsFarmer
			stats.CurrentStreak = max(1, stats.CurrentStreak+1)
		} else {
			stats.Losses++
			scoreChange = LoseAsFarmer
			stats.CurrentStreak = min(-1, stats.CurrentStreak-1)
		}
	}

	// 连胜加成
	if stats.CurrentStreak >= 10 {
		scoreChange += StreakBonus10
	} else if stats.CurrentStreak >= 5 {
		scoreChange += StreakBonus5
	} else if stats.CurrentStreak >= 3 {
		scoreChange += StreakBonus3
	}

	// 更新最大连胜
	if stats.CurrentStreak > stats.MaxWinStreak {
		stats.MaxWinStreak = stats.CurrentStreak
	}

	// 更新积分（最低为0）
	stats.Score = max(0, stats.Score+scoreChange)

	// 保存统计
	if err := lm.SavePlayerStats(ctx, stats); err != nil {
		return err
	}

	// 更新排行榜
	if err := lm.UpdateLeaderboard(ctx, stats); err != nil {
		return err
	}

	return nil
}

// UpdateLeaderboard 更新排行榜
func (lm *LeaderboardManager) UpdateLeaderboard(ctx context.Context, stats *PlayerStats) error {
	// 更新总排行榜
	if err := lm.redis.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  float64(stats.Score),
		Member: stats.PlayerID,
	}).Err(); err != nil {
		return err
	}

	// 更新每日排行榜
	today := time.Now().Format("2006-01-02")
	dailyKey := dailyLeaderboard + today
	if err := lm.redis.ZAdd(ctx, dailyKey, redis.Z{
		Score:  float64(stats.Score),
		Member: stats.PlayerID,
	}).Err(); err != nil {
		return err
	}
	// 设置过期时间（2天）
	lm.redis.Expire(ctx, dailyKey, 48*time.Hour)

	// 更新每周排行榜
	year, week := time.Now().ISOWeek()
	weeklyKey := fmt.Sprintf("%s%d-W%02d", weeklyLeaderboard, year, week)
	if err := lm.redis.ZAdd(ctx, weeklyKey, redis.Z{
		Score:  float64(stats.Score),
		Member: stats.PlayerID,
	}).Err(); err != nil {
		return err
	}
	// 设置过期时间（8天）
	lm.redis.Expire(ctx, weeklyKey, 8*24*time.Hour)

	return nil
}

// GetLeaderboard 获取排行榜
func (lm *LeaderboardManager) GetLeaderboard(ctx context.Context, leaderboardType string, offset, limit int) ([]LeaderboardEntry, error) {
	// 确定使用哪个排行榜
	key := leaderboardKey
	switch leaderboardType {
	case "daily":
		today := time.Now().Format("2006-01-02")
		key = dailyLeaderboard + today
	case "weekly":
		year, week := time.Now().ISOWeek()
		key = fmt.Sprintf("%s%d-W%02d", weeklyLeaderboard, year, week)
	}

	// 获取排行榜（从高到低）
	results, err := lm.redis.ZRevRangeWithScores(ctx, key, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, 0, len(results))
	for i, result := range results {
		playerID := result.Member.(string)

		// 获取玩家详细统计
		stats, err := lm.GetPlayerStats(ctx, playerID)
		if err != nil || stats == nil {
			continue
		}

		winRate := 0.0
		if stats.TotalGames > 0 {
			winRate = float64(stats.Wins) / float64(stats.TotalGames) * 100
		}

		entries = append(entries, LeaderboardEntry{
			Rank:       offset + i + 1,
			PlayerID:   playerID,
			PlayerName: stats.PlayerName,
			Score:      int(result.Score),
			Wins:       stats.Wins,
			WinRate:    winRate,
		})
	}

	return entries, nil
}

// GetPlayerRank 获取玩家排名
func (lm *LeaderboardManager) GetPlayerRank(ctx context.Context, playerID string) (int64, error) {
	rank, err := lm.redis.ZRevRank(ctx, leaderboardKey, playerID).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil // 未上榜
		}
		return -1, err
	}
	return rank + 1, nil // Redis 排名从 0 开始
}

// GetTopPlayers 获取前 N 名玩家（带详细信息）
func (lm *LeaderboardManager) GetTopPlayers(ctx context.Context, n int) ([]LeaderboardEntry, error) {
	return lm.GetLeaderboard(ctx, "total", 0, n)
}

// GetAroundPlayer 获取玩家附近的排名
func (lm *LeaderboardManager) GetAroundPlayer(ctx context.Context, playerID string, count int) ([]LeaderboardEntry, error) {
	rank, err := lm.GetPlayerRank(ctx, playerID)
	if err != nil || rank == -1 {
		return nil, err
	}

	// 计算偏移量
	offset := int(rank) - count/2 - 1
	if offset < 0 {
		offset = 0
	}

	return lm.GetLeaderboard(ctx, "total", offset, count)
}

// --- 辅助函数 ---

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SortByScore 按积分排序
func SortByScore(entries []LeaderboardEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})
}

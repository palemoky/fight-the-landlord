package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis key 前缀
	roomKeyPrefix    = "room:"
	sessionKeyPrefix = "session:"
	matchQueueKey    = "match:queue"

	// 房间数据过期时间
	roomExpiration = 2 * time.Hour
)

// game.RoomData 房间数据（用于 Redis 序列化）
type RoomData struct {
	Code        string           `json:"code"`
	State       int              `json:"state"`
	Players     []PlayerData     `json:"players"`
	PlayerOrder []string         `json:"player_order"`
	CreatedAt   int64            `json:"created_at"`
	GameData    *GameSessionData `json:"game_data,omitempty"`
}

// PlayerData 玩家数据
type PlayerData struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Seat       int    `json:"seat"`
	Ready      bool   `json:"ready"`
	IsLandlord bool   `json:"is_landlord"`
}

// game.GameSessionData 游戏会话数据（简化版，用于恢复基本信息）
type GameSessionData struct {
	State         int     `json:"state"`
	CurrentPlayer int     `json:"current_player"`
	LandlordIdx   int     `json:"landlord_idx"`
	PlayerHands   [][]int `json:"player_hands"` // 简化为点数列表
	BottomCards   []int   `json:"bottom_cards"`
}

// RedisStore Redis 存储
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore 创建 Redis 存储
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// --- 房间存储 ---

// SaveRoom 保存房间到 Redis
func (rs *RedisStore) SaveRoom(ctx context.Context, roomCode string, data *RoomData) error {
	if data == nil {
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化房间数据失败: %w", err)
	}

	key := roomKeyPrefix + roomCode
	return rs.client.Set(ctx, key, jsonData, roomExpiration).Err()
}

// Loadgame.Room 从 Redis 加载房间（仅返回数据，需要外部重建）
func (rs *RedisStore) LoadRoom(ctx context.Context, code string) (*RoomData, error) {
	key := roomKeyPrefix + code
	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // 房间不存在
		}
		return nil, err
	}

	var roomData RoomData
	if err := json.Unmarshal(data, &roomData); err != nil {
		return nil, fmt.Errorf("反序列化房间数据失败: %w", err)
	}

	return &roomData, nil
}

// Deletegame.Room 从 Redis 删除房间
func (rs *RedisStore) DeleteRoom(ctx context.Context, code string) error {
	key := roomKeyPrefix + code
	return rs.client.Del(ctx, key).Err()
}

// GetAllgame.RoomCodes 获取所有房间号
func (rs *RedisStore) GetAllRoomCodes(ctx context.Context) ([]string, error) {
	keys, err := rs.client.Keys(ctx, roomKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	codes := make([]string, len(keys))
	for i, key := range keys {
		codes[i] = key[len(roomKeyPrefix):]
	}
	return codes, nil
}

// --- 匹配队列 ---

// AddToMatchQueue 添加玩家到匹配队列
func (rs *RedisStore) AddToMatchQueue(ctx context.Context, playerID string) error {
	return rs.client.RPush(ctx, matchQueueKey, playerID).Err()
}

// RemoveFromMatchQueue 从匹配队列移除玩家
func (rs *RedisStore) RemoveFromMatchQueue(ctx context.Context, playerID string) error {
	return rs.client.LRem(ctx, matchQueueKey, 0, playerID).Err()
}

// GetMatchQueueLength 获取匹配队列长度
func (rs *RedisStore) GetMatchQueueLength(ctx context.Context) (int64, error) {
	return rs.client.LLen(ctx, matchQueueKey).Result()
}

// PopFromMatchQueue 从匹配队列弹出指定数量的玩家
func (rs *RedisStore) PopFromMatchQueue(ctx context.Context, count int) ([]string, error) {
	pipe := rs.client.Pipeline()
	results := make([]*redis.StringCmd, count)

	for i := 0; i < count; i++ {
		results[i] = pipe.LPop(ctx, matchQueueKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	players := make([]string, 0, count)
	for _, result := range results {
		if playerID, err := result.Result(); err == nil {
			players = append(players, playerID)
		}
	}

	return players, nil
}

// --- 会话存储 ---

// PlayerSessionData 玩家会话数据（用于 Redis 序列化）
type PlayerSessionData struct {
	PlayerID       string `json:"player_id"`
	PlayerName     string `json:"player_name"`
	ReconnectToken string `json:"token"`
	RoomCode       string `json:"room_code"`
	IsOnline       bool   `json:"is_online"`
	DisconnectedAt int64  `json:"disconnected_at,omitempty"`
}

// SaveSession 保存会话到 Redis
func (rs *RedisStore) SaveSession(ctx context.Context, session *PlayerSessionData) error {
	data := map[string]any{
		"player_id":   session.PlayerID,
		"player_name": session.PlayerName,
		"token":       session.ReconnectToken,
		"room_code":   session.RoomCode,
		"is_online":   session.IsOnline,
	}

	if session.DisconnectedAt != 0 {
		data["disconnected_at"] = session.DisconnectedAt
	}

	key := sessionKeyPrefix + session.PlayerID
	return rs.client.HSet(ctx, key, data).Err()
}

// LoadSession 从 Redis 加载会话
func (rs *RedisStore) LoadSession(ctx context.Context, playerID string) (*PlayerSessionData, error) {
	key := sessionKeyPrefix + playerID
	data, err := rs.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	session := &PlayerSessionData{
		PlayerID:       data["player_id"],
		PlayerName:     data["player_name"],
		ReconnectToken: data["token"],
		RoomCode:       data["room_code"],
		IsOnline:       data["is_online"] == "1",
	}

	return session, nil
}

// DeleteSession 删除会话
func (rs *RedisStore) DeleteSession(ctx context.Context, playerID string) error {
	key := sessionKeyPrefix + playerID
	return rs.client.Del(ctx, key).Err()
}

// --- 辅助方法 ---

// SetRoomExpiration 设置房间过期时间
func (rs *RedisStore) SetRoomExpiration(ctx context.Context, code string, expiration time.Duration) error {
	key := roomKeyPrefix + code
	return rs.client.Expire(ctx, key, expiration).Err()
}

package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Game rules configuration
const (
	// BottomCardsPublic controls whether bottom cards are visible to all players
	// true: All players see bottom cards, all counters deduct them
	// false: Only landlord sees bottom cards, only landlord's counter deducts them
	BottomCardsPublic = true
)

// 默认配置值
const (
	defaultHost                  = "0.0.0.0"
	defaultPort                  = 1780
	defaultMaxConnections        = 10000
	defaultRedisAddr             = "localhost:6379"
	defaultTurnTimeout           = 30
	defaultBidTimeout            = 15
	defaultRoomTimeout           = 10
	defaultShutdownTimeout       = 30
	defaultShutdownCheckInterval = 15
	defaultRoomCleanupDelay      = 30
	defaultRateLimitPerSecond    = 10
	defaultRateLimitPerMinute    = 60
	defaultBanDuration           = 60
	defaultMessageLimitPerSecond = 20
	defaultChatLimitPerSecond    = 1
	defaultChatLimitPerMinute    = 30
	defaultChatCooldown          = 5
)

// Config 服务端配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Redis    RedisConfig    `yaml:"redis"`
	Game     GameConfig     `yaml:"game"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig WebSocket 服务器配置
type ServerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	MaxConnections int    `yaml:"max_connections"` // 最大并发连接数，0 表示无限制
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GameConfig 游戏配置
type GameConfig struct {
	TurnTimeout           int `yaml:"turn_timeout"`            // 出牌超时（秒）
	BidTimeout            int `yaml:"bid_timeout"`             // 叫地主超时（秒）
	RoomTimeout           int `yaml:"room_timeout"`            // 房间等待超时（分钟）
	ShutdownTimeout       int `yaml:"shutdown_timeout"`        // 优雅关闭超时（分钟）
	ShutdownCheckInterval int `yaml:"shutdown_check_interval"` // 优雅关闭检测间隔（秒）
	RoomCleanupDelay      int `yaml:"room_cleanup_delay"`      // 游戏结束后房间清理延迟（秒）
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AllowedOrigins []string           `yaml:"allowed_origins"` // 允许的来源
	RateLimit      RateLimitConfig    `yaml:"rate_limit"`      // 连接速率限制
	MessageLimit   MessageLimitConfig `yaml:"message_limit"`   // 消息速率限制
	ChatLimit      ChatLimitConfig    `yaml:"chat_limit"`      // 聊天消息速率限制
}

// RateLimitConfig 连接速率限制配置
type RateLimitConfig struct {
	MaxPerSecond int `yaml:"max_per_second"` // 每秒最大连接数
	MaxPerMinute int `yaml:"max_per_minute"` // 每分钟最大连接数
	BanDuration  int `yaml:"ban_duration"`   // 封禁时长（秒）
}

// MessageLimitConfig 消息速率限制配置
type MessageLimitConfig struct {
	MaxPerSecond int `yaml:"max_per_second"` // 每秒最大消息数
}

// ChatLimitConfig 聊天消息速率限制配置
type ChatLimitConfig struct {
	MaxPerSecond int `yaml:"max_per_second"` // 每秒最大聊天消息数
	MaxPerMinute int `yaml:"max_per_minute"` // 每分钟最大聊天消息数
	Cooldown     int `yaml:"cooldown"`       // 冷却时间（秒）
}

// Duration 方法
func (c *GameConfig) TurnTimeoutDuration() time.Duration {
	return time.Duration(c.TurnTimeout) * time.Second
}

func (c *GameConfig) BidTimeoutDuration() time.Duration {
	return time.Duration(c.BidTimeout) * time.Second
}

func (c *GameConfig) RoomTimeoutDuration() time.Duration {
	return time.Duration(c.RoomTimeout) * time.Minute
}

func (c *GameConfig) ShutdownTimeoutDuration() time.Duration {
	return time.Duration(c.ShutdownTimeout) * time.Minute
}

func (c *GameConfig) ShutdownCheckIntervalDuration() time.Duration {
	return time.Duration(c.ShutdownCheckInterval) * time.Second
}

func (c *GameConfig) RoomCleanupDelayDuration() time.Duration {
	return time.Duration(c.RoomCleanupDelay) * time.Second
}

func (c *RateLimitConfig) BanDurationTime() time.Duration {
	return time.Duration(c.BanDuration) * time.Second
}

func (c *ChatLimitConfig) CooldownDuration() time.Duration {
	return time.Duration(c.Cooldown) * time.Second
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	loadFromEnv(&cfg)

	return &cfg, nil
}

// --- 环境变量辅助函数 ---

func getEnvStr(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

func getEnvInt(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}

func getEnvStrSlice(key string, target *[]string) {
	if v := os.Getenv(key); v != "" {
		*target = strings.Split(v, ",")
	}
}

// loadFromEnv 从环境变量加载配置（覆盖文件配置）
func loadFromEnv(cfg *Config) {
	// Server
	getEnvStr("SERVER_HOST", &cfg.Server.Host)
	getEnvInt("SERVER_PORT", &cfg.Server.Port)
	getEnvInt("SERVER_MAX_CONNECTIONS", &cfg.Server.MaxConnections)

	// Redis
	getEnvStr("REDIS_ADDR", &cfg.Redis.Addr)
	getEnvStr("REDIS_PASSWORD", &cfg.Redis.Password)
	getEnvInt("REDIS_DB", &cfg.Redis.DB)

	// Game
	getEnvInt("GAME_TURN_TIMEOUT", &cfg.Game.TurnTimeout)
	getEnvInt("GAME_BID_TIMEOUT", &cfg.Game.BidTimeout)
	getEnvInt("GAME_ROOM_TIMEOUT", &cfg.Game.RoomTimeout)
	getEnvInt("GAME_SHUTDOWN_TIMEOUT", &cfg.Game.ShutdownTimeout)
	getEnvInt("GAME_SHUTDOWN_CHECK_INTERVAL", &cfg.Game.ShutdownCheckInterval)
	getEnvInt("GAME_ROOM_CLEANUP_DELAY", &cfg.Game.RoomCleanupDelay)

	// Security
	getEnvStrSlice("SECURITY_ALLOWED_ORIGINS", &cfg.Security.AllowedOrigins)
	getEnvInt("SECURITY_RATE_LIMIT_PER_SECOND", &cfg.Security.RateLimit.MaxPerSecond)
	getEnvInt("SECURITY_MESSAGE_LIMIT_PER_SECOND", &cfg.Security.MessageLimit.MaxPerSecond)
}

// --- 默认值辅助函数 ---

func setDefaultStr(target *string, defaultVal string) {
	if *target == "" {
		*target = defaultVal
	}
}

func setDefaultInt(target *int, defaultVal int) {
	if *target == 0 {
		*target = defaultVal
	}
}

func setDefaultStrSlice(target *[]string, defaultVal []string) {
	if len(*target) == 0 {
		*target = defaultVal
	}
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	// Server
	setDefaultStr(&cfg.Server.Host, defaultHost)
	setDefaultInt(&cfg.Server.Port, defaultPort)
	setDefaultInt(&cfg.Server.MaxConnections, defaultMaxConnections)

	// Redis
	setDefaultStr(&cfg.Redis.Addr, defaultRedisAddr)

	// Game
	setDefaultInt(&cfg.Game.TurnTimeout, defaultTurnTimeout)
	setDefaultInt(&cfg.Game.BidTimeout, defaultBidTimeout)
	setDefaultInt(&cfg.Game.RoomTimeout, defaultRoomTimeout)
	setDefaultInt(&cfg.Game.ShutdownTimeout, defaultShutdownTimeout)
	setDefaultInt(&cfg.Game.ShutdownCheckInterval, defaultShutdownCheckInterval)
	setDefaultInt(&cfg.Game.RoomCleanupDelay, defaultRoomCleanupDelay)

	// Security
	setDefaultStrSlice(&cfg.Security.AllowedOrigins, []string{"*"})
	setDefaultInt(&cfg.Security.RateLimit.MaxPerSecond, defaultRateLimitPerSecond)
	setDefaultInt(&cfg.Security.RateLimit.MaxPerMinute, defaultRateLimitPerMinute)
	setDefaultInt(&cfg.Security.RateLimit.BanDuration, defaultBanDuration)
	setDefaultInt(&cfg.Security.MessageLimit.MaxPerSecond, defaultMessageLimitPerSecond)
	setDefaultInt(&cfg.Security.ChatLimit.MaxPerSecond, defaultChatLimitPerSecond)
	setDefaultInt(&cfg.Security.ChatLimit.MaxPerMinute, defaultChatLimitPerMinute)
	setDefaultInt(&cfg.Security.ChatLimit.Cooldown, defaultChatCooldown)
}

// Default 返回默认配置
func Default() *Config {
	// 尝试加载默认配置文件
	if cfg, err := Load("configs/config.yaml"); err == nil {
		return cfg
	} else {
		log.Printf("无法加载默认配置文件，使用最小默认值: %v", err)
	}

	// 使用 setDefaults 设置的默认值
	cfg := &Config{}
	setDefaults(cfg)
	return cfg
}

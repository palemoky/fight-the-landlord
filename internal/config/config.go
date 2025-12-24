package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
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
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GameConfig 游戏配置
type GameConfig struct {
	TurnTimeout int `yaml:"turn_timeout"` // 出牌超时（秒）
	BidTimeout  int `yaml:"bid_timeout"`  // 叫地主超时（秒）
	RoomTimeout int `yaml:"room_timeout"` // 房间等待超时（分钟）
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AllowedOrigins []string           `yaml:"allowed_origins"` // 允许的来源
	RateLimit      RateLimitConfig    `yaml:"rate_limit"`      // 连接速率限制
	MessageLimit   MessageLimitConfig `yaml:"message_limit"`   // 消息速率限制
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

// TurnTimeoutDuration 返回出牌超时时长
func (c *GameConfig) TurnTimeoutDuration() time.Duration {
	return time.Duration(c.TurnTimeout) * time.Second
}

// BidTimeoutDuration 返回叫地主超时时长
func (c *GameConfig) BidTimeoutDuration() time.Duration {
	return time.Duration(c.BidTimeout) * time.Second
}

// RoomTimeoutDuration 返回房间等待超时时长
func (c *GameConfig) RoomTimeoutDuration() time.Duration {
	return time.Duration(c.RoomTimeout) * time.Minute
}

// BanDurationTime 返回封禁时长
func (c *RateLimitConfig) BanDurationTime() time.Duration {
	return time.Duration(c.BanDuration) * time.Second
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

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 1780
	}
	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "localhost:6379"
	}
	if cfg.Game.TurnTimeout == 0 {
		cfg.Game.TurnTimeout = 30
	}
	if cfg.Game.BidTimeout == 0 {
		cfg.Game.BidTimeout = 15
	}
	if cfg.Game.RoomTimeout == 0 {
		cfg.Game.RoomTimeout = 10
	}
	// 安全配置默认值
	if len(cfg.Security.AllowedOrigins) == 0 {
		cfg.Security.AllowedOrigins = []string{"*"}
	}
	if cfg.Security.RateLimit.MaxPerSecond == 0 {
		cfg.Security.RateLimit.MaxPerSecond = 10
	}
	if cfg.Security.RateLimit.MaxPerMinute == 0 {
		cfg.Security.RateLimit.MaxPerMinute = 60
	}
	if cfg.Security.RateLimit.BanDuration == 0 {
		cfg.Security.RateLimit.BanDuration = 60
	}
	if cfg.Security.MessageLimit.MaxPerSecond == 0 {
		cfg.Security.MessageLimit.MaxPerSecond = 20
	}
}

// Default 返回默认配置
func Default() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 1780,
		},
		Redis: RedisConfig{
			Addr: "localhost:6379",
		},
		Game: GameConfig{
			TurnTimeout: 30,
			BidTimeout:  15,
			RoomTimeout: 10,
		},
		Security: SecurityConfig{
			AllowedOrigins: []string{"*"},
			RateLimit: RateLimitConfig{
				MaxPerSecond: 10,
				MaxPerMinute: 60,
				BanDuration:  60,
			},
			MessageLimit: MessageLimitConfig{
				MaxPerSecond: 20,
			},
		},
	}
	return cfg
}

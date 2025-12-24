package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 服务端配置
type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
	Game   GameConfig   `yaml:"game"`
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

	return &cfg, nil
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
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
	}
}

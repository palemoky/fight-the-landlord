package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	t.Parallel()

	// Create a temp config file
	content := `
server:
  host: "127.0.0.1"
  port: 8080
  max_connections: 5000

redis:
  addr: "redis:6379"
  password: "secret"
  db: 1

game:
  turn_timeout: 60
  bid_timeout: 30
  room_timeout: 15
  shutdown_timeout: 60
  shutdown_check_interval: 30
  room_cleanup_delay: 60

security:
  allowed_origins:
    - "http://localhost:3000"
    - "https://example.com"
  rate_limit:
    max_per_second: 20
    max_per_minute: 120
    ban_duration: 120
  message_limit:
    max_per_second: 50
  chat_limit:
    max_per_second: 2
    max_per_minute: 60
    cooldown: 10
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0o600)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify loaded values
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 5000, cfg.Server.MaxConnections)
	assert.Equal(t, "redis:6379", cfg.Redis.Addr)
	assert.Equal(t, "secret", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, 60, cfg.Game.TurnTimeout)
	assert.Equal(t, 30, cfg.Game.BidTimeout)
	assert.Len(t, cfg.Security.AllowedOrigins, 2)
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Parallel()

	cfg, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: :::"), 0o600)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_AppliesDefaults(t *testing.T) {
	t.Parallel()

	// Empty config file - defaults should be applied
	content := `{}`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.yaml")
	err := os.WriteFile(configPath, []byte(content), 0o600)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify defaults are applied
	assert.Equal(t, defaultHost, cfg.Server.Host)
	assert.Equal(t, defaultPort, cfg.Server.Port)
	assert.Equal(t, defaultMaxConnections, cfg.Server.MaxConnections)
	assert.Equal(t, defaultRedisAddr, cfg.Redis.Addr)
	assert.Equal(t, defaultTurnTimeout, cfg.Game.TurnTimeout)
	assert.Equal(t, defaultBidTimeout, cfg.Game.BidTimeout)
	assert.Equal(t, []string{"*"}, cfg.Security.AllowedOrigins)
}

func TestDefault(t *testing.T) {
	// Note: Not parallel because Default() reads from filesystem

	cfg := Default()
	require.NotNil(t, cfg)

	// Verify defaults are set
	assert.Equal(t, defaultHost, cfg.Server.Host)
	assert.Equal(t, defaultPort, cfg.Server.Port)
	assert.Equal(t, defaultTurnTimeout, cfg.Game.TurnTimeout)
}

func TestGameConfig_DurationMethods(t *testing.T) {
	t.Parallel()

	cfg := &GameConfig{
		TurnTimeout:           30,
		BidTimeout:            15,
		RoomTimeout:           10,
		ShutdownTimeout:       60,
		ShutdownCheckInterval: 5,
		RoomCleanupDelay:      20,
	}

	assert.Equal(t, 30*time.Second, cfg.TurnTimeoutDuration())
	assert.Equal(t, 15*time.Second, cfg.BidTimeoutDuration())
	assert.Equal(t, 10*time.Minute, cfg.RoomTimeoutDuration())
	assert.Equal(t, 60*time.Minute, cfg.ShutdownTimeoutDuration())
	assert.Equal(t, 5*time.Second, cfg.ShutdownCheckIntervalDuration())
	assert.Equal(t, 20*time.Second, cfg.RoomCleanupDelayDuration())
}

func TestRateLimitConfig_BanDurationTime(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{BanDuration: 120}
	assert.Equal(t, 120*time.Second, cfg.BanDurationTime())
}

func TestChatLimitConfig_CooldownDuration(t *testing.T) {
	t.Parallel()

	cfg := &ChatLimitConfig{Cooldown: 10}
	assert.Equal(t, 10*time.Second, cfg.CooldownDuration())
}

func TestLoadFromEnv(t *testing.T) {
	// Not parallel because it modifies environment variables

	// Set environment variables
	t.Setenv("SERVER_HOST", "env-host")
	t.Setenv("SERVER_PORT", "9999")
	t.Setenv("REDIS_ADDR", "env-redis:6380")
	t.Setenv("GAME_TURN_TIMEOUT", "120")
	t.Setenv("SECURITY_ALLOWED_ORIGINS", "http://a.com,http://b.com")

	// Create minimal config file
	content := `{}`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "env.yaml")
	err := os.WriteFile(configPath, []byte(content), 0o600)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify env vars override defaults
	assert.Equal(t, "env-host", cfg.Server.Host)
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "env-redis:6380", cfg.Redis.Addr)
	assert.Equal(t, 120, cfg.Game.TurnTimeout)
	assert.Equal(t, []string{"http://a.com", "http://b.com"}, cfg.Security.AllowedOrigins)
}

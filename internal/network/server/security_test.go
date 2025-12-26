package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	// 5 reqs/sec, 10 reqs/min, 1s ban
	rl := NewRateLimiter(5, 10, 1*time.Second)
	ip := "127.0.0.1"

	// Initial requests should be allowed
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(ip), "Request %d should be allowed", i)
	}

	// 6th request should fail due to per-second limit
	assert.False(t, rl.Allow(ip), "6th request should be blocked")
	assert.True(t, rl.IsBanned(ip), "IP should be banned")
}

func TestIPFilter(t *testing.T) {
	filter := NewIPFilter()
	ip := "192.168.1.1"

	// Default allow
	assert.True(t, filter.IsAllowed(ip))

	// Blacklist
	filter.AddToBlacklist(ip)
	assert.False(t, filter.IsAllowed(ip))

	filter.RemoveFromBlacklist(ip)
	assert.True(t, filter.IsAllowed(ip))

	// Whitelist (only whitelist allowed if present)
	filter.AddToWhitelist("10.0.0.1")
	assert.False(t, filter.IsAllowed(ip))
	assert.True(t, filter.IsAllowed("10.0.0.1"))
}

func TestMessageRateLimiter(t *testing.T) {
	// 5 msgs/sec
	ml := NewMessageRateLimiter(5)
	clientID := "client1"

	// Allowed
	for i := 0; i < 5; i++ {
		allowed, warning := ml.AllowMessage(clientID)
		assert.True(t, allowed)
		// Warning threshold is usually max/2 = 2. So after 2nd (count 3, 4, 5) it might warn.
		// Implementation check: warningThreshold = maxPerSecond / 2 = 2.
		// If count > 2, warning = true.
		if i >= 3 {
			assert.True(t, warning, "Should warn after threshold")
		}
	}

	// 6th message should be blocked
	allowed, warning := ml.AllowMessage(clientID)
	assert.False(t, allowed)
	assert.True(t, warning)
}

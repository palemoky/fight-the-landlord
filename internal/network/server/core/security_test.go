package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	tests := []struct {
		name    string
		ip      string
		setup   func(*IPFilter)
		allowed bool
	}{
		{
			name:    "Default allow",
			ip:      "192.168.1.1",
			setup:   func(f *IPFilter) {},
			allowed: true,
		},
		{
			name: "Blacklisted IP",
			ip:   "192.168.1.2",
			setup: func(f *IPFilter) {
				f.AddToBlacklist("192.168.1.2")
			},
			allowed: false,
		},
		{
			name: "Removed from blacklist",
			ip:   "192.168.1.3",
			setup: func(f *IPFilter) {
				f.AddToBlacklist("192.168.1.3")
				f.RemoveFromBlacklist("192.168.1.3")
			},
			allowed: true,
		},
		{
			name: "Whitelist enforcement (IP not in whitelist)",
			ip:   "192.168.1.4",
			setup: func(f *IPFilter) {
				f.AddToWhitelist("10.0.0.1")
			},
			allowed: false,
		},
		{
			name: "Whitelist enforcement (IP in whitelist)",
			ip:   "10.0.0.1",
			setup: func(f *IPFilter) {
				f.AddToWhitelist("10.0.0.1")
			},
			allowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := NewIPFilter()
			if tt.setup != nil {
				tt.setup(f)
			}
			assert.Equal(t, tt.allowed, f.IsAllowed(tt.ip))
		})
	}
}

func TestMessageRateLimiter(t *testing.T) {
	t.Parallel()

	// 5 msgs/sec
	ml := NewMessageRateLimiter(5)
	clientID := "client1"

	// Allowed
	for i := range 5 {
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

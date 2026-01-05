package server

import (
	"net/http"
	"sync"
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

func TestRateLimiter_BurstTraffic(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(10, 50, 2*time.Second)
	ip := "192.168.1.1"

	// Simulate burst traffic
	for i := 0; i < 10; i++ {
		assert.True(t, rl.Allow(ip), "Burst request %d should be allowed", i)
	}

	// 11th request should be blocked
	assert.False(t, rl.Allow(ip))
	assert.True(t, rl.IsBanned(ip))

	// Wait for ban to expire
	time.Sleep(2100 * time.Millisecond)

	// Should be allowed again
	assert.False(t, rl.IsBanned(ip))
	assert.True(t, rl.Allow(ip))
}

func TestRateLimiter_MinuteLimit(t *testing.T) {
	t.Parallel()

	// 100/sec but only 5/min
	rl := NewRateLimiter(100, 5, 1*time.Second)
	ip := "10.0.0.1"

	// First 5 requests allowed
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(ip))
	}

	// 6th request blocked by minute limit
	assert.False(t, rl.Allow(ip))
}

func TestRateLimiter_Concurrency(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(100, 200, 1*time.Second)
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Concurrent requests from same IP
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow("concurrent-test") {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	assert.Greater(t, successCount, 0)
	assert.LessOrEqual(t, successCount, 50)
}

func TestChatRateLimiter_CooldownPeriod(t *testing.T) {
	t.Parallel()

	// 2/sec, 5/min, 3s cooldown
	cl := NewChatRateLimiter(2, 5, 3*time.Second)
	clientID := "chatter1"

	// First 2 messages allowed
	allowed, reason := cl.AllowChat(clientID)
	assert.True(t, allowed)
	assert.Empty(t, reason)

	allowed, reason = cl.AllowChat(clientID)
	assert.True(t, allowed)
	assert.Empty(t, reason)

	// 3rd message triggers cooldown
	allowed, reason = cl.AllowChat(clientID)
	assert.False(t, allowed)
	assert.Contains(t, reason, "派大星")

	// During cooldown, messages blocked
	allowed, reason = cl.AllowChat(clientID)
	assert.False(t, allowed)
	assert.Contains(t, reason, "章鱼哥")

	// Wait for cooldown to expire
	time.Sleep(3100 * time.Millisecond)

	// Should be allowed again
	allowed, reason = cl.AllowChat(clientID)
	assert.True(t, allowed)
	assert.Empty(t, reason)
}

func TestChatRateLimiter_MinuteLimit(t *testing.T) {
	t.Parallel()

	// 10/sec, 3/min, 2s cooldown
	cl := NewChatRateLimiter(10, 3, 2*time.Second)
	clientID := "spammer"

	// First 3 messages allowed
	for i := 0; i < 3; i++ {
		allowed, _ := cl.AllowChat(clientID)
		assert.True(t, allowed, "Message %d should be allowed", i)
	}

	// 4th message blocked by minute limit
	allowed, reason := cl.AllowChat(clientID)
	assert.False(t, allowed)
	assert.Contains(t, reason, "休息")
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
		{
			name: "Blacklist overrides whitelist",
			ip:   "10.0.0.2",
			setup: func(f *IPFilter) {
				f.AddToWhitelist("10.0.0.2")
				f.AddToBlacklist("10.0.0.2")
			},
			allowed: false,
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

func TestGetClientIP_ProxyHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "Direct connection",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 10.0.0.2, 10.0.0.3",
			},
			expectedIP: "203.0.113.1", // First IP is the original client
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			expectedIP: "203.0.113.2",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.3",
				"X-Real-IP":       "203.0.113.4",
			},
			expectedIP: "203.0.113.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := GetClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
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

func TestMessageRateLimiter_WarningCount(t *testing.T) {
	t.Parallel()

	ml := NewMessageRateLimiter(3)
	clientID := "test-client"

	// Trigger warnings
	for i := 0; i < 5; i++ {
		ml.AllowMessage(clientID)
	}

	// Check warning count
	warnings := ml.GetWarningCount(clientID)
	assert.Greater(t, warnings, 0)
}

func TestMessageRateLimiter_ClearRateLimit(t *testing.T) {
	t.Parallel()

	ml := NewMessageRateLimiter(5)
	clientID := "temp-client"

	// Generate some activity
	ml.AllowMessage(clientID)
	ml.AllowMessage(clientID)

	// Remove client
	ml.ClearRateLimit(clientID)

	// Should start fresh
	allowed, warning := ml.AllowMessage(clientID)
	assert.True(t, allowed)
	assert.False(t, warning)
}

func TestOriginChecker_AllowAll(t *testing.T) {
	t.Parallel()

	oc := NewOriginChecker([]string{"*"})
	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Origin", "https://evil.com")

	assert.True(t, oc.Check(req))
}

func TestOriginChecker_SpecificOrigins(t *testing.T) {
	t.Parallel()

	oc := NewOriginChecker([]string{"https://example.com", "https://app.example.com"})

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"https://example.com", true},
		{"https://app.example.com", true},
		{"https://evil.com", false},
		{"http://example.com", false}, // Different scheme
		{"", true},                    // No origin header (same-origin or local)
	}

	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
		if tt.origin != "" {
			req.Header.Set("Origin", tt.origin)
		}
		assert.Equal(t, tt.allowed, oc.Check(req), "Origin: %s", tt.origin)
	}
}

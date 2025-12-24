package server

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	requests map[string]*clientRate
	mu       sync.RWMutex

	// 配置
	maxRequestsPerSecond int           // 每秒最大请求数
	maxRequestsPerMinute int           // 每分钟最大请求数
	banDuration          time.Duration // 封禁时长
	cleanupInterval      time.Duration // 清理间隔
}

// clientRate 客户端速率记录
type clientRate struct {
	secondCount int       // 当前秒请求数
	minuteCount int       // 当前分钟请求数
	lastSecond  time.Time // 上次秒级计数时间
	lastMinute  time.Time // 上次分钟计数时间
	bannedUntil time.Time // 封禁到期时间
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(maxPerSecond, maxPerMinute int, banDuration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:             make(map[string]*clientRate),
		maxRequestsPerSecond: maxPerSecond,
		maxRequestsPerMinute: maxPerMinute,
		banDuration:          banDuration,
		cleanupInterval:      5 * time.Minute,
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rate, exists := rl.requests[ip]

	if !exists {
		rl.requests[ip] = &clientRate{
			secondCount: 1,
			minuteCount: 1,
			lastSecond:  now,
			lastMinute:  now,
		}
		return true
	}

	// 检查是否被封禁
	if now.Before(rate.bannedUntil) {
		return false
	}

	// 重置秒级计数
	if now.Sub(rate.lastSecond) >= time.Second {
		rate.secondCount = 0
		rate.lastSecond = now
	}

	// 重置分钟计数
	if now.Sub(rate.lastMinute) >= time.Minute {
		rate.minuteCount = 0
		rate.lastMinute = now
	}

	rate.secondCount++
	rate.minuteCount++

	// 检查是否超限
	if rate.secondCount > rl.maxRequestsPerSecond || rate.minuteCount > rl.maxRequestsPerMinute {
		rate.bannedUntil = now.Add(rl.banDuration)
		log.Printf("⚠️ IP %s 因请求过于频繁被暂时封禁 %v", ip, rl.banDuration)
		return false
	}

	return true
}

// IsBanned 检查 IP 是否被封禁
func (rl *RateLimiter) IsBanned(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rate, exists := rl.requests[ip]
	if !exists {
		return false
	}

	return time.Now().Before(rate.bannedUntil)
}

// cleanup 清理过期记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, rate := range rl.requests {
			// 如果超过 10 分钟没有请求，删除记录
			if now.Sub(rate.lastMinute) > 10*time.Minute && now.After(rate.bannedUntil) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// --- 来源验证 ---

// OriginChecker 来源验证器
type OriginChecker struct {
	allowedOrigins map[string]bool
	allowAll       bool
}

// NewOriginChecker 创建来源验证器
func NewOriginChecker(origins []string) *OriginChecker {
	oc := &OriginChecker{
		allowedOrigins: make(map[string]bool),
	}

	for _, origin := range origins {
		if origin == "*" {
			oc.allowAll = true
			return oc
		}
		oc.allowedOrigins[strings.ToLower(origin)] = true
	}

	return oc
}

// Check 检查来源是否允许
func (oc *OriginChecker) Check(r *http.Request) bool {
	if oc.allowAll {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// 没有 Origin 头，可能是同源请求或本地客户端
		return true
	}

	return oc.allowedOrigins[strings.ToLower(origin)]
}

// --- IP 白名单/黑名单 ---

// IPFilter IP 过滤器
type IPFilter struct {
	whitelist map[string]bool // 白名单
	blacklist map[string]bool // 黑名单
	mu        sync.RWMutex
}

// NewIPFilter 创建 IP 过滤器
func NewIPFilter() *IPFilter {
	return &IPFilter{
		whitelist: make(map[string]bool),
		blacklist: make(map[string]bool),
	}
}

// AddToWhitelist 添加到白名单
func (f *IPFilter) AddToWhitelist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.whitelist[ip] = true
}

// AddToBlacklist 添加到黑名单
func (f *IPFilter) AddToBlacklist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blacklist[ip] = true
}

// RemoveFromBlacklist 从黑名单移除
func (f *IPFilter) RemoveFromBlacklist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.blacklist, ip)
}

// IsAllowed 检查 IP 是否允许
func (f *IPFilter) IsAllowed(ip string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 如果有白名单且不在白名单中，拒绝
	if len(f.whitelist) > 0 && !f.whitelist[ip] {
		return false
	}

	// 如果在黑名单中，拒绝
	if f.blacklist[ip] {
		return false
	}

	return true
}

// --- 辅助函数 ---

// GetClientIP 获取客户端真实 IP
func GetClientIP(r *http.Request) string {
	// 检查代理头
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// 取第一个 IP（最原始的客户端）
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 从连接中获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// --- 消息速率限制 ---

// MessageRateLimiter 消息速率限制器（针对已连接的客户端）
type MessageRateLimiter struct {
	limits map[string]*messageRate
	mu     sync.RWMutex

	maxMessagesPerSecond int
	warningThreshold     int // 警告阈值
}

type messageRate struct {
	count     int
	lastReset time.Time
	warnings  int // 警告次数
}

// NewMessageRateLimiter 创建消息速率限制器
func NewMessageRateLimiter(maxPerSecond int) *MessageRateLimiter {
	return &MessageRateLimiter{
		limits:               make(map[string]*messageRate),
		maxMessagesPerSecond: maxPerSecond,
		warningThreshold:     maxPerSecond / 2,
	}
}

// AllowMessage 检查是否允许发送消息
func (ml *MessageRateLimiter) AllowMessage(clientID string) (allowed bool, warning bool) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	now := time.Now()
	rate, exists := ml.limits[clientID]

	if !exists {
		ml.limits[clientID] = &messageRate{
			count:     1,
			lastReset: now,
		}
		return true, false
	}

	// 如果超过 1 秒，重置计数
	if now.Sub(rate.lastReset) >= time.Second {
		rate.count = 1
		rate.lastReset = now
		return true, false
	}

	rate.count++

	// 超过限制
	if rate.count > ml.maxMessagesPerSecond {
		rate.warnings++
		return false, true
	}

	// 接近限制，发出警告
	if rate.count > ml.warningThreshold {
		return true, true
	}

	return true, false
}

// GetWarningCount 获取警告次数
func (ml *MessageRateLimiter) GetWarningCount(clientID string) int {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	rate, exists := ml.limits[clientID]
	if !exists {
		return 0
	}
	return rate.warnings
}

// RemoveClient 移除客户端记录
func (ml *MessageRateLimiter) RemoveClient(clientID string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	delete(ml.limits, clientID)
}

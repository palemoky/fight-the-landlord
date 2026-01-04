package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	// 重连等待时间
	reconnectTimeout = 2 * time.Minute
	// 会话过期时间
	sessionExpireTime = 10 * time.Minute
)

// PlayerSession 玩家会话（用于断线重连）
type PlayerSession struct {
	PlayerID       string
	PlayerName     string
	ReconnectToken string
	RoomCode       string

	DisconnectedAt time.Time // 断线时间
	IsOnline       bool      // 是否在线

	mu sync.RWMutex
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*PlayerSession // playerID -> session
	tokens   map[string]string         // token -> playerID
	mu       sync.RWMutex
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*PlayerSession),
		tokens:   make(map[string]string),
	}

	// 启动会话清理协程
	go sm.cleanupLoop()

	return sm
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(playerID, playerName string) *PlayerSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token := generateToken()

	session := &PlayerSession{
		PlayerID:       playerID,
		PlayerName:     playerName,
		ReconnectToken: token,
		IsOnline:       true,
	}

	sm.sessions[playerID] = session
	sm.tokens[token] = playerID

	return session
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(playerID string) *PlayerSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[playerID]
}

// GetSessionByToken 通过 token 获取会话
func (sm *SessionManager) GetSessionByToken(token string) *PlayerSession {
	sm.mu.RLock()
	playerID, ok := sm.tokens[token]
	if !ok {
		sm.mu.RUnlock()
		return nil
	}
	session := sm.sessions[playerID]
	sm.mu.RUnlock()
	return session
}

// SetOffline 设置玩家离线
func (sm *SessionManager) SetOffline(playerID string) {
	sm.mu.RLock()
	session, ok := sm.sessions[playerID]
	sm.mu.RUnlock()

	if ok {
		session.mu.Lock()
		session.IsOnline = false
		session.DisconnectedAt = time.Now()
		session.mu.Unlock()
	}
}

// SetOnline 设置玩家上线
func (sm *SessionManager) SetOnline(playerID string) {
	sm.mu.RLock()
	session, ok := sm.sessions[playerID]
	sm.mu.RUnlock()

	if ok {
		session.mu.Lock()
		session.IsOnline = true
		session.DisconnectedAt = time.Time{}
		session.mu.Unlock()
	}
}

// SetRoom 设置玩家所在房间
func (sm *SessionManager) SetRoom(playerID, roomCode string) {
	sm.mu.RLock()
	session, ok := sm.sessions[playerID]
	sm.mu.RUnlock()

	if ok {
		session.mu.Lock()
		session.RoomCode = roomCode
		session.mu.Unlock()
	}
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(playerID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, ok := sm.sessions[playerID]; ok {
		delete(sm.tokens, session.ReconnectToken)
		delete(sm.sessions, playerID)
	}
}

// CanReconnect 检查玩家是否可以重连
func (sm *SessionManager) CanReconnect(token, playerID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	storedPlayerID, ok := sm.tokens[token]
	if !ok || storedPlayerID != playerID {
		return false
	}

	session, ok := sm.sessions[playerID]
	if !ok {
		return false
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// 检查是否在重连时限内
	if !session.IsOnline && time.Since(session.DisconnectedAt) > reconnectTimeout {
		return false
	}

	return true
}

// IsOnline 检查玩家是否在线
func (sm *SessionManager) IsOnline(playerID string) bool {
	sm.mu.RLock()
	session, ok := sm.sessions[playerID]
	sm.mu.RUnlock()

	if !ok {
		return false
	}

	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.IsOnline
}

// cleanupLoop 定期清理过期会话
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup 清理过期会话
func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for playerID, session := range sm.sessions {
		session.mu.RLock()
		// 清理离线超过会话过期时间的会话
		if !session.IsOnline && now.Sub(session.DisconnectedAt) > sessionExpireTime {
			delete(sm.tokens, session.ReconnectToken)
			delete(sm.sessions, playerID)
		}
		session.mu.RUnlock()
	}
}

// generateToken 生成随机 token
func generateToken() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

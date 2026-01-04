package handler

import (
	"log"
	"sync"

	"github.com/palemoky/fight-the-landlord/internal/game/match"
	"github.com/palemoky/fight-the-landlord/internal/game/room"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/server/session"
	"github.com/palemoky/fight-the-landlord/internal/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/types"
)

// HandlerDeps 处理器依赖
type HandlerDeps struct {
	Server         types.ServerInterface
	RoomManager    *room.RoomManager
	Matcher        *match.Matcher
	ChatLimiter    types.ChatLimiter
	Leaderboard    *storage.LeaderboardManager
	SessionManager *session.SessionManager
}

// Handler 消息处理器
type Handler struct {
	server         types.ServerInterface
	roomManager    *room.RoomManager
	matcher        *match.Matcher
	chatLimiter    types.ChatLimiter
	leaderboard    *storage.LeaderboardManager
	sessionManager *session.SessionManager
	handlers       map[protocol.MessageType]handlerFunc
	games          map[string]*session.GameSession
	gamesMu        sync.RWMutex
}

// handlerFunc 统一的处理器函数签名
type handlerFunc func(client types.ClientInterface, msg *protocol.Message)

// NewHandler 创建处理器
func NewHandler(deps HandlerDeps) *Handler {
	h := &Handler{
		server:         deps.Server,
		roomManager:    deps.RoomManager,
		matcher:        deps.Matcher,
		chatLimiter:    deps.ChatLimiter,
		leaderboard:    deps.Leaderboard,
		sessionManager: deps.SessionManager,
		games:          make(map[string]*session.GameSession),
	}
	h.initHandlers()
	return h
}

// GetGameSession 获取房间的游戏会话
func (h *Handler) GetGameSession(roomCode string) *session.GameSession {
	h.gamesMu.RLock()
	defer h.gamesMu.RUnlock()
	return h.games[roomCode]
}

// SetGameSession 设置房间的游戏会话
func (h *Handler) SetGameSession(roomCode string, gs *session.GameSession) {
	h.gamesMu.Lock()
	defer h.gamesMu.Unlock()
	if gs == nil {
		delete(h.games, roomCode)
	} else {
		h.games[roomCode] = gs
	}
}

// initHandlers 初始化消息处理器映射
func (h *Handler) initHandlers() {
	h.handlers = map[protocol.MessageType]handlerFunc{
		// 连接操作
		protocol.MsgPing:      h.handlePing,
		protocol.MsgReconnect: h.handleReconnect,

		// 房间操作
		protocol.MsgCreateRoom:  func(c types.ClientInterface, _ *protocol.Message) { h.handleCreateRoom(c) },
		protocol.MsgJoinRoom:    h.handleJoinRoom,
		protocol.MsgLeaveRoom:   func(c types.ClientInterface, _ *protocol.Message) { h.handleLeaveRoom(c) },
		protocol.MsgQuickMatch:  func(c types.ClientInterface, _ *protocol.Message) { h.handleQuickMatch(c) },
		protocol.MsgReady:       func(c types.ClientInterface, _ *protocol.Message) { h.handleReady(c, true) },
		protocol.MsgCancelReady: func(c types.ClientInterface, _ *protocol.Message) { h.handleReady(c, false) },

		// 游戏操作
		protocol.MsgBid:       h.handleBid,
		protocol.MsgPlayCards: h.handlePlayCards,
		protocol.MsgPass:      func(c types.ClientInterface, _ *protocol.Message) { h.handlePass(c) },

		// 信息查询
		protocol.MsgGetStats:             func(c types.ClientInterface, _ *protocol.Message) { h.handleGetStats(c) },
		protocol.MsgGetLeaderboard:       h.handleGetLeaderboard,
		protocol.MsgGetRoomList:          func(c types.ClientInterface, _ *protocol.Message) { h.handleGetRoomList(c) },
		protocol.MsgGetOnlineCount:       func(c types.ClientInterface, _ *protocol.Message) { h.handleGetOnlineCount(c) },
		protocol.MsgGetMaintenanceStatus: func(c types.ClientInterface, _ *protocol.Message) { h.handleGetMaintenanceStatus(c) },
		protocol.MsgChat:                 h.handleChat,
	}
}

// Handle 处理消息
func (h *Handler) Handle(client types.ClientInterface, msg *protocol.Message) {
	if handler, ok := h.handlers[msg.Type]; ok {
		handler(client, msg)
		return
	}

	log.Printf("⚠️  未知消息类型: '%s' (来自玩家: %s, ID: %s)", msg.Type, client.GetName(), client.GetID())
	log.Printf("    消息详情: Payload长度=%d bytes", len(msg.Payload))
	client.SendMessage(codec.NewErrorMessage(protocol.ErrCodeInvalidMsg))
}

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/palemoky/fight-the-landlord/internal/config"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/server/core"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game"
	"github.com/palemoky/fight-the-landlord/internal/network/server/game/session"
	"github.com/palemoky/fight-the-landlord/internal/network/server/handlers"
	"github.com/palemoky/fight-the-landlord/internal/network/server/storage"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境需要限制
	},
	// 启用 permessage-deflate 压缩扩展
	// 可减少 40-70% 流量，gorilla/websocket 会自动协商压缩参数
	// 压缩会对CPU和内存造成压力，只有在大文件压缩才有收益，大量小文件反而是负优化
	EnableCompression: false,
}

// Server WebSocket 服务器
type Server struct {
	config         *config.Config
	redis          *redis.Client
	redisStore     *storage.RedisStore
	leaderboard    *storage.LeaderboardManager
	roomManager    *game.RoomManager
	matcher        *game.Matcher
	sessionManager *session.SessionManager
	clients        map[string]*Client
	clientsMu      sync.RWMutex
	handler        *handlers.Handler

	// 安全组件
	rateLimiter    *core.RateLimiter
	originChecker  *core.OriginChecker
	messageLimiter *core.MessageRateLimiter
	chatLimiter    *core.ChatRateLimiter
	ipFilter       *core.IPFilter

	// 连接控制
	maxConnections int
	semaphore      chan struct{} // 信号量控制并发连接数

	// 维护模式
	maintenanceMode bool
	maintenanceMu   sync.RWMutex
}

// NewServer 创建服务器实例
func NewServer(cfg *config.Config) (*Server, error) {
	// 初始化 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis 连接失败: %w", err)
	}

	s := &Server{
		config:         cfg,
		redis:          rdb,
		redisStore:     storage.NewRedisStore(rdb),
		leaderboard:    storage.NewLeaderboardManager(rdb),
		clients:        make(map[string]*Client),
		sessionManager: session.NewSessionManager(),
		// 初始化安全组件
		rateLimiter: core.NewRateLimiter(
			cfg.Security.RateLimit.MaxPerSecond,
			cfg.Security.RateLimit.MaxPerMinute,
			cfg.Security.RateLimit.BanDurationTime(),
		),
		originChecker:  core.NewOriginChecker(cfg.Security.AllowedOrigins),
		messageLimiter: core.NewMessageRateLimiter(cfg.Security.MessageLimit.MaxPerSecond),
		chatLimiter: core.NewChatRateLimiter(
			cfg.Security.ChatLimit.MaxPerSecond,
			cfg.Security.ChatLimit.MaxPerMinute,
			cfg.Security.ChatLimit.CooldownDuration(),
		),
		ipFilter: core.NewIPFilter(),
		// 初始化连接控制
		maxConnections: cfg.Server.MaxConnections,
		semaphore:      make(chan struct{}, cfg.Server.MaxConnections),
	}

	// 初始化房间管理器
	s.roomManager = game.NewRoomManager(s)

	// 初始化匹配器
	s.matcher = game.NewMatcher(s)

	// 初始化消息处理器
	s.handler = handlers.NewHandler(s)

	log.Printf("🔒 安全配置: 连接限制=%d/s, 消息限制=%d/s, 聊天限制=%d/s, 最大连接数=%d",
		cfg.Security.RateLimit.MaxPerSecond, cfg.Security.MessageLimit.MaxPerSecond, cfg.Security.ChatLimit.MaxPerSecond, cfg.Server.MaxConnections)

	return s, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/health", s.handleHealth)

	// 启动监控 goroutine
	go s.monitorStats()

	log.Printf("🚀 服务器启动在 ws://%s/ws (CPU核心数: %d)", addr, runtime.NumCPU())
	server := &http.Server{
		Addr:              addr,
		Handler:           nil,
		ReadHeaderTimeout: 10 * time.Second, // 防止 Slowloris 攻击
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return server.ListenAndServe()
}

// handleWebSocket 处理 WebSocket 连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 获取真实客户端IP
	clientIP := core.GetClientIP(r)

	// 维护模式检查（最优先）
	if s.IsMaintenanceMode() {
		log.Printf("🔧 维护模式，拒绝新连接: %s", clientIP)
		http.Error(w, "Server is under maintenance, please try again later",
			http.StatusServiceUnavailable)
		return
	}

	// 连接数限制检查
	select {
	case s.semaphore <- struct{}{}:
		// 成功获取信号量，连接建立后释放
		defer func() { <-s.semaphore }()
	default:
		log.Printf("🚫 达到最大连接数限制 (%d), IP: %s", s.maxConnections, clientIP)
		http.Error(w, "Server Full", http.StatusServiceUnavailable)
		return
	}

	// IP 过滤检查
	if !s.ipFilter.IsAllowed(clientIP) {
		log.Printf("🚫 IP %s 被过滤器拒绝", clientIP)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 来源验证
	if !s.originChecker.Check(r) {
		log.Printf("🚫 来源验证失败: %s (IP: %s)", r.Header.Get("Origin"), clientIP)
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	// 速率限制检查
	if !s.rateLimiter.Allow(clientIP) {
		log.Printf("🚫 IP %s 请求过于频繁", clientIP)
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	// 创建客户端
	client := NewClient(s, conn)
	client.IP = clientIP // 记录客户端 IP
	s.registerClient(client)

	// 创建会话
	session := s.sessionManager.CreateSession(client.ID, client.Name)

	// 发送连接成功消息（包含重连令牌）
	client.SendMessage(codec.MustNewMessage(protocol.MsgConnected, protocol.ConnectedPayload{
		PlayerID:       client.ID,
		PlayerName:     client.Name,
		ReconnectToken: session.ReconnectToken,
	}))

	log.Printf("✅ 玩家 %s (%s) 已连接", client.Name, client.ID)

	// 启动客户端读写协程
	go client.ReadPump()
	go client.WritePump()
}

// handleHealth 健康检查接口
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// registerClient 注册客户端
func (s *Server) registerClient(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[client.ID] = client
}

// unregisterClient 注销客户端
func (s *Server) unregisterClient(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if _, ok := s.clients[client.ID]; ok {
		delete(s.clients, client.ID)
		log.Printf("❌ 玩家 %s (%s) 已断开", client.Name, client.ID)
	}
}

// GetOnlineCount 获取在线人数（按需调用）
func (s *Server) GetOnlineCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// Broadcast 广播消息给所有客户端
func (s *Server) Broadcast(msg *protocol.Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		client.SendMessage(msg)
	}
}

// BroadcastToLobby 广播消息给大厅玩家（未在房间内的玩家）
func (s *Server) BroadcastToLobby(msg *protocol.Message) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.GetRoom() == "" {
			client.SendMessage(msg)
		}
	}
}

// monitorStats 定期监控服务器状态
func (s *Server) monitorStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		onlineCount := s.GetOnlineCount()
		goroutines := runtime.NumGoroutine()
		activeConns := len(s.semaphore)

		log.Printf("📊 [监控] 在线: %d | Goroutines: %d | 活跃连接: %d/%d | 内存: %.2f MB",
			onlineCount,
			goroutines,
			activeConns,
			s.maxConnections,
			float64(m.Alloc)/1024/1024)
	}
}

// EnterMaintenanceMode 进入维护模式
func (s *Server) EnterMaintenanceMode() {
	s.maintenanceMu.Lock()
	s.maintenanceMode = true
	s.maintenanceMu.Unlock()

	// 通知大厅用户服务器即将关闭
	s.BroadcastToLobby(codec.MustNewMessage(protocol.MsgError, protocol.ErrorPayload{
		Code:    protocol.ErrCodeServerMaintenance,
		Message: "👷🏻‍♂️ 维护模式：停止新的房间创建",
	}))

	log.Println("🔧 进入维护模式：停止新连接和房间创建")
}

// IsMaintenanceMode 检查是否在维护模式
func (s *Server) IsMaintenanceMode() bool {
	s.maintenanceMu.RLock()
	defer s.maintenanceMu.RUnlock()
	return s.maintenanceMode
}

// GracefulShutdown 优雅关闭服务器
func (s *Server) GracefulShutdown(timeout time.Duration) {
	// 1. 进入维护模式
	s.EnterMaintenanceMode()

	// 2. 等待游戏结束
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(s.config.Game.ShutdownCheckIntervalDuration())
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		activeGames := s.roomManager.GetActiveGamesCount()
		if activeGames == 0 {
			log.Printf("✅ 所有房间已结束，将在 %ds 后关闭服务器！\n", s.config.Game.RoomCleanupDelay)

			// 通知大厅用户服务器即将关闭
			s.BroadcastToLobby(codec.MustNewMessage(protocol.MsgError, protocol.ErrorPayload{
				Code:    protocol.ErrCodeServerMaintenance,
				Message: fmt.Sprintf("🚧 服务器将在 %d 秒后停机维护！", s.config.Game.RoomCleanupDelay),
			}))

			break
		}
		log.Printf("⏳ 等待 %d 个房间结束...", activeGames)
		<-ticker.C
	}

	// 3. 超时检查
	if activeGames := s.roomManager.GetActiveGamesCount(); activeGames > 0 {
		log.Printf("⚠️ 超时，仍有 %d 个房间进行中，强制关闭", activeGames)
	}

	// 4. 发送通知（如果配置了）
	s.sendShutdownNotification()

	// 5. 关闭服务器
	s.Shutdown()
}

// sendShutdownNotification 发送关闭通知到小米音箱
func (s *Server) sendShutdownNotification() {
	// 从环境变量读取小米音箱配置
	speakerURL := os.Getenv("XIAOMI_SPEAKER_URL")
	if speakerURL == "" {
		return // 未配置，跳过
	}

	message := "斗地主服务器已优雅关闭，开始升级吧！"

	// 发送 POST 请求
	payload := fmt.Sprintf(`{"text":"%s"}`, message)
	req, err := http.NewRequest("POST", speakerURL, strings.NewReader(payload))
	if err != nil {
		log.Printf("创建通知请求失败: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	// 添加认证 Headers（如果配置了）
	if apiSecret := os.Getenv("XIAOMI_SPEAKER_API_SECRET"); apiSecret != "" {
		req.Header.Set("Speaker-API-Secret", apiSecret)
	}
	if cfClientID := os.Getenv("XIAOMI_SPEAKER_CF_CLIENT_ID"); cfClientID != "" {
		req.Header.Set("CF-Access-Client-Id", cfClientID)
	}
	if cfClientSecret := os.Getenv("XIAOMI_SPEAKER_CF_CLIENT_SECRET"); cfClientSecret != "" {
		req.Header.Set("CF-Access-Client-Secret", cfClientSecret)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送通知失败: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		log.Println("🔔 已发送关闭通知到小米音箱")
	} else {
		log.Printf("通知响应异常: %d", resp.StatusCode)
	}
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	time.Sleep(s.config.Game.RoomCleanupDelayDuration())

	// 关闭所有客户端连接
	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMu.Unlock()

	// 关闭 Redis
	_ = s.redis.Close()

	log.Println("服务器已关闭")
}

// Interface implementations for types.ServerContext
func (s *Server) GetRedisStore() types.RedisStoreInterface         { return s.redisStore }
func (s *Server) GetLeaderboard() types.LeaderboardInterface       { return s.leaderboard }
func (s *Server) GetSessionManager() types.SessionManagerInterface { return s.sessionManager }
func (s *Server) GetRoomManager() types.RoomManagerInterface       { return s.roomManager }
func (s *Server) GetMatcher() types.MatcherInterface               { return s.matcher }
func (s *Server) GetGameConfig() types.GameConfigInterface         { return &s.config.Game }
func (s *Server) GetChatLimiter() types.ChatLimiterInterface       { return s.chatLimiter }

func (s *Server) GetClientByID(id string) types.ClientInterface {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.clients[id]
}

func (s *Server) RegisterClient(id string, client types.ClientInterface) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	if c, ok := client.(*Client); ok {
		s.clients[id] = c
	}
}

func (s *Server) UnregisterClient(id string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	delete(s.clients, id)
}

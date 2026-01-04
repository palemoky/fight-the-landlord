package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
)

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
	payloadData := map[string]string{"text": message}
	payloadBytes, _ := json.Marshal(payloadData)
	req, err := http.NewRequest(http.MethodPost, speakerURL, bytes.NewReader(payloadBytes))
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

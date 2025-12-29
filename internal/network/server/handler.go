package server

import (
	"context"
	"log"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// Handler 消息处理器
type Handler struct {
	server *Server
}

// NewHandler 创建处理器
func NewHandler(s *Server) *Handler {
	return &Handler{server: s}
}

// Handle 处理消息
func (h *Handler) Handle(client *Client, msg *protocol.Message) {
	switch msg.Type {
	// 连接操作
	case protocol.MsgPing:
		h.handlePing(client, msg)
	case protocol.MsgReconnect:
		h.handleReconnect(client, msg)

	// 房间操作
	case protocol.MsgCreateRoom:
		h.handleCreateRoom(client)
	case protocol.MsgJoinRoom:
		h.handleJoinRoom(client, msg)
	case protocol.MsgLeaveRoom:
		h.handleLeaveRoom(client)
	case protocol.MsgQuickMatch:
		h.handleQuickMatch(client)
	case protocol.MsgReady:
		h.handleReady(client, true)
	case protocol.MsgCancelReady:
		h.handleReady(client, false)

	// 游戏操作
	case protocol.MsgBid:
		h.handleBid(client, msg)
	case protocol.MsgPlayCards:
		h.handlePlayCards(client, msg)
	case protocol.MsgPass:
		h.handlePass(client)

	// 排行榜操作
	case protocol.MsgGetStats:
		h.handleGetStats(client)
	case protocol.MsgGetLeaderboard:
		h.handleGetLeaderboard(client, msg)
	case protocol.MsgGetRoomList:
		h.handleGetRoomList(client)
	case protocol.MsgGetOnlineCount:
		h.handleGetOnlineCount(client)
	case protocol.MsgGetMaintenanceStatus:
		h.handleGetMaintenanceStatus(client)
	case protocol.MsgChat:
		h.handleChat(client, msg)

	default:
		log.Printf("⚠️  未知消息类型: '%s' (来自玩家: %s, ID: %s)", msg.Type, client.Name, client.ID)
		log.Printf("    消息详情: Payload长度=%d bytes", len(msg.Payload))
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeInvalidMsg))
	}
}

// handlePing 处理心跳消息
func (h *Handler) handlePing(client *Client, msg *protocol.Message) {
	payload, err := protocol.ParsePayload[protocol.PingPayload](msg)
	if err != nil {
		return
	}

	// 立即回复 pong
	client.SendMessage(protocol.MustNewMessage(protocol.MsgPong, protocol.PongPayload{
		ClientTimestamp: payload.Timestamp,
		ServerTimestamp: time.Now().UnixMilli(),
	}))
}

// handleReconnect 处理断线重连
func (h *Handler) handleReconnect(client *Client, msg *protocol.Message) {
	payload, err := protocol.ParsePayload[protocol.ReconnectPayload](msg)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	// 验证重连令牌
	if !h.server.sessionManager.CanReconnect(payload.Token, payload.PlayerID) {
		client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, "重连令牌无效或已过期"))
		return
	}

	// 获取旧会话
	session := h.server.sessionManager.GetSession(payload.PlayerID)
	if session == nil {
		client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, "会话不存在"))
		return
	}

	// 更新客户端 ID 和名称为原来的
	oldID := client.ID
	client.ID = session.PlayerID
	client.Name = session.PlayerName

	// 从旧 ID 注销，用新 ID 注册
	h.server.clientsMu.Lock()
	delete(h.server.clients, oldID)
	h.server.clients[client.ID] = client
	h.server.clientsMu.Unlock()

	// 标记会话上线
	h.server.sessionManager.SetOnline(client.ID)

	// 构建重连响应
	reconnectPayload := protocol.ReconnectedPayload{
		PlayerID:   client.ID,
		PlayerName: client.Name,
	}

	// 如果在房间中，恢复房间信息
	if session.RoomCode != "" {
		room := h.server.roomManager.GetRoom(session.RoomCode)
		if room != nil {
			// 更新客户端在房间中的引用
			room.mu.Lock()
			if player, ok := room.Players[client.ID]; ok {
				player.Client = client
			}
			room.mu.Unlock()

			client.SetRoom(session.RoomCode)
			reconnectPayload.RoomCode = session.RoomCode

			// 通知其他玩家该玩家已重连
			room.mu.RLock()
			for id, p := range room.Players {
				if id != client.ID && p.Client != nil {
					p.Client.SendMessage(protocol.MustNewMessage(protocol.MsgPlayerOnline, protocol.PlayerOnlinePayload{
						PlayerID:   client.ID,
						PlayerName: client.Name,
					}))
				}
			}
			room.mu.RUnlock()

			// 如果游戏正在进行，恢复游戏状态
			game := room.GetGameSession()
			if game != nil {
				reconnectPayload.GameState = h.buildGameStateDTO(game, client.ID)
			}
		}
	}

	// 发送重连成功消息
	client.SendMessage(protocol.MustNewMessage(protocol.MsgReconnected, reconnectPayload))

	log.Printf("🔄 玩家 %s (%s) 重连成功", client.Name, client.ID)
}

// buildGameStateDTO 构建游戏状态 DTO
func (h *Handler) buildGameStateDTO(game *GameSession, playerID string) *protocol.GameStateDTO {
	game.mu.RLock()
	defer game.mu.RUnlock()

	// 查找玩家的手牌
	var hand []protocol.CardInfo
	for _, p := range game.players {
		if p.ID == playerID {
			hand = protocol.CardsToInfos(p.Hand)
			break
		}
	}

	// 构建玩家信息列表
	players := make([]protocol.PlayerInfo, len(game.players))
	for i, p := range game.players {
		players[i] = protocol.PlayerInfo{
			ID:         p.ID,
			Name:       p.Name,
			Seat:       p.Seat,
			IsLandlord: p.IsLandlord,
			CardsCount: len(p.Hand),
			Online:     h.server.sessionManager.IsOnline(p.ID),
		}
	}

	// 确定游戏阶段
	phase := "waiting"
	switch game.state {
	case GameStateBidding:
		phase = "bidding"
	case GameStatePlaying:
		phase = "playing"
	case GameStateEnded:
		phase = "ended"
	}

	// 当前回合玩家 ID
	currentTurnID := ""
	switch game.state {
	case GameStateBidding:
		currentTurnID = game.players[game.currentBidder].ID
	case GameStatePlaying:
		currentTurnID = game.players[game.currentPlayer].ID
	}

	// 上家出的牌
	var lastPlayed []protocol.CardInfo
	lastPlayerID := ""
	if !game.lastPlayedHand.IsEmpty() {
		lastPlayed = protocol.CardsToInfos(game.lastPlayedHand.Cards)
		lastPlayerID = game.players[game.lastPlayerIdx].ID
	}

	return &protocol.GameStateDTO{
		Phase:         phase,
		Players:       players,
		Hand:          hand,
		LandlordCards: protocol.CardsToInfos(game.bottomCards),
		CurrentTurn:   currentTurnID,
		LastPlayed:    lastPlayed,
		LastPlayerID:  lastPlayerID,
		MustPlay:      game.lastPlayerIdx == game.currentPlayer || game.lastPlayedHand.IsEmpty(),
		CanBeat:       true, // 简化处理
	}
}

// handleCreateRoom 处理创建房间
func (h *Handler) handleCreateRoom(client *Client) {
	// 维护模式检查
	if h.server.IsMaintenanceMode() {
		client.SendMessage(protocol.NewErrorMessageWithText(
			protocol.ErrCodeServerMaintenance, "服务器维护中，暂停创建房间"))
		return
	}

	// 如果已在房间中，先离开
	if client.GetRoom() != "" {
		h.server.roomManager.LeaveRoom(client)
	}

	room, err := h.server.roomManager.CreateRoom(client)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		return
	}

	client.SendMessage(protocol.MustNewMessage(protocol.MsgRoomCreated, protocol.RoomCreatedPayload{
		RoomCode: room.Code,
		Player:   room.getPlayerInfo(client.ID),
	}))
}

// handleJoinRoom 处理加入房间
func (h *Handler) handleJoinRoom(client *Client, msg *protocol.Message) {
	// 维护模式检查
	if h.server.IsMaintenanceMode() {
		client.SendMessage(protocol.NewErrorMessageWithText(
			protocol.ErrCodeServerMaintenance, "服务器维护中，暂停加入房间"))
		return
	}

	payload, err := protocol.ParsePayload[protocol.JoinRoomPayload](msg)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	// 如果已在房间中，先离开
	if client.GetRoom() != "" {
		h.server.roomManager.LeaveRoom(client)
	}

	room, err := h.server.roomManager.JoinRoom(client, payload.RoomCode)
	if err != nil {
		if roomErr, ok := err.(*RoomError); ok {
			client.SendMessage(protocol.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
		return
	}

	client.SendMessage(protocol.MustNewMessage(protocol.MsgRoomJoined, protocol.RoomJoinedPayload{
		RoomCode: room.Code,
		Player:   room.getPlayerInfo(client.ID),
		Players:  room.getAllPlayersInfo(),
	}))
}

// handleLeaveRoom 处理离开房间
func (h *Handler) handleLeaveRoom(client *Client) {
	h.server.roomManager.LeaveRoom(client)
}

// handleQuickMatch 处理快速匹配
func (h *Handler) handleQuickMatch(client *Client) {
	// 维护模式检查
	if h.server.IsMaintenanceMode() {
		client.SendMessage(protocol.NewErrorMessageWithText(
			protocol.ErrCodeServerMaintenance, "服务器维护中，暂停快速匹配"))
		return
	}

	// 如果已在房间中，先离开
	if client.GetRoom() != "" {
		h.server.roomManager.LeaveRoom(client)
	}

	h.server.matcher.AddToQueue(client)
}

// handleReady 处理准备
func (h *Handler) handleReady(client *Client, ready bool) {
	err := h.server.roomManager.SetPlayerReady(client, ready)
	if err != nil {
		if roomErr, ok := err.(*RoomError); ok {
			client.SendMessage(protocol.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// handleBid 处理叫地主
func (h *Handler) handleBid(client *Client, msg *protocol.Message) {
	payload, err := protocol.ParsePayload[protocol.BidPayload](msg)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	room := h.server.roomManager.GetRoom(client.GetRoom())
	if room == nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	game := room.GetGameSession()
	if game == nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := game.HandleBid(client.ID, payload.Bid); err != nil {
		if roomErr, ok := err.(*RoomError); ok {
			client.SendMessage(protocol.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// handlePlayCards 处理出牌
func (h *Handler) handlePlayCards(client *Client, msg *protocol.Message) {
	payload, err := protocol.ParsePayload[protocol.PlayCardsPayload](msg)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeInvalidMsg))
		return
	}

	room := h.server.roomManager.GetRoom(client.GetRoom())
	if room == nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	game := room.GetGameSession()
	if game == nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := game.HandlePlayCards(client.ID, payload.Cards); err != nil {
		if roomErr, ok := err.(*RoomError); ok {
			client.SendMessage(protocol.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// handlePass 处理不出
func (h *Handler) handlePass(client *Client) {
	room := h.server.roomManager.GetRoom(client.GetRoom())
	if room == nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeNotInRoom))
		return
	}

	game := room.GetGameSession()
	if game == nil {
		client.SendMessage(protocol.NewErrorMessage(protocol.ErrCodeGameNotStart))
		return
	}

	if err := game.HandlePass(client.ID); err != nil {
		if roomErr, ok := err.(*RoomError); ok {
			client.SendMessage(protocol.NewErrorMessage(roomErr.Code))
		} else {
			client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, err.Error()))
		}
	}
}

// --- 排行榜处理 ---

// handleGetStats 获取个人统计
func (h *Handler) handleGetStats(client *Client) {
	ctx := context.Background()
	stats, err := h.server.leaderboard.GetPlayerStats(ctx, client.ID)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, "获取统计失败"))
		return
	}

	if stats == nil {
		// 没有统计数据，返回空数据
		client.SendMessage(protocol.MustNewMessage(protocol.MsgStatsResult, protocol.StatsResultPayload{
			PlayerID:   client.ID,
			PlayerName: client.Name,
		}))
		return
	}

	// 获取排名
	rank, _ := h.server.leaderboard.GetPlayerRank(ctx, client.ID)

	winRate := 0.0
	if stats.TotalGames > 0 {
		winRate = float64(stats.Wins) / float64(stats.TotalGames) * 100
	}

	client.SendMessage(protocol.MustNewMessage(protocol.MsgStatsResult, protocol.StatsResultPayload{
		PlayerID:      stats.PlayerID,
		PlayerName:    stats.PlayerName,
		TotalGames:    stats.TotalGames,
		Wins:          stats.Wins,
		Losses:        stats.Losses,
		WinRate:       winRate,
		LandlordGames: stats.LandlordGames,
		LandlordWins:  stats.LandlordWins,
		FarmerGames:   stats.FarmerGames,
		FarmerWins:    stats.FarmerWins,
		Score:         stats.Score,
		Rank:          int(rank),
		CurrentStreak: stats.CurrentStreak,
		MaxWinStreak:  stats.MaxWinStreak,
	}))
}

// handleGetLeaderboard 获取排行榜
func (h *Handler) handleGetLeaderboard(client *Client, msg *protocol.Message) {
	payload, err := protocol.ParsePayload[protocol.GetLeaderboardPayload](msg)
	if err != nil {
		// 默认获取总排行榜前 10
		payload = &protocol.GetLeaderboardPayload{
			Type:   "total",
			Offset: 0,
			Limit:  10,
		}
	}

	// 限制请求数量
	if payload.Limit <= 0 || payload.Limit > 50 {
		payload.Limit = 10
	}
	if payload.Offset < 0 {
		payload.Offset = 0
	}

	entries, err := h.server.leaderboard.GetLeaderboard(context.Background(), payload.Type, payload.Offset, payload.Limit)
	if err != nil {
		client.SendMessage(protocol.NewErrorMessageWithText(protocol.ErrCodeUnknown, "获取排行榜失败"))
		return
	}

	// 转换为协议格式
	protocolEntries := make([]protocol.LeaderboardEntry, len(entries))
	for i, e := range entries {
		protocolEntries[i] = protocol.LeaderboardEntry{
			Rank:       e.Rank,
			PlayerID:   e.PlayerID,
			PlayerName: e.PlayerName,
			Score:      e.Score,
			Wins:       e.Wins,
			WinRate:    e.WinRate,
		}
	}

	client.SendMessage(protocol.MustNewMessage(protocol.MsgLeaderboardResult, protocol.LeaderboardResultPayload{
		Type:    payload.Type,
		Entries: protocolEntries,
	}))
}

// handleGetRoomList 获取房间列表
func (h *Handler) handleGetRoomList(client *Client) {
	rooms := h.server.roomManager.GetRoomList()

	client.SendMessage(protocol.MustNewMessage(protocol.MsgRoomListResult, protocol.RoomListResultPayload{
		Rooms: rooms,
	}))
}

// handleGetOnlineCount 获取在线人数（按需）
func (h *Handler) handleGetOnlineCount(client *Client) {
	count := h.server.GetOnlineCount()

	client.SendMessage(protocol.MustNewMessage(protocol.MsgOnlineCount, protocol.OnlineCountPayload{
		Count: count,
	}))
}

// handleGetMaintenanceStatus 获取维护状态
func (h *Handler) handleGetMaintenanceStatus(client *Client) {
	maintenance := h.server.IsMaintenanceMode()

	client.SendMessage(protocol.MustNewMessage(protocol.MsgMaintenanceStatus, protocol.MaintenanceStatusPayload{
		Maintenance: maintenance,
	}))
}

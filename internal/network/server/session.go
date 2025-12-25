package server

import (
	"context"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/rule"
)

// GameState 游戏状态
type GameState int

const (
	GameStateInit GameState = iota
	GameStateBidding
	GameStatePlaying
	GameStateEnded
)

const (
	// 玩家离线等待时间（秒）
	offlineWaitTimeout = 30 * time.Second
)

// GamePlayer 游戏中的玩家
type GamePlayer struct {
	ID         string
	Name       string
	Seat       int
	Hand       []card.Card
	IsLandlord bool
	IsOffline  bool // 是否离线
}

// GameSession 游戏会话
type GameSession struct {
	room    *Room
	state   GameState
	players []*GamePlayer // 按座位顺序

	deck          card.Deck
	landlordCards []card.Card

	// 叫地主相关
	currentBidder int // 当前叫地主的玩家索引
	highestBidder int // 叫地主的玩家索引，-1 表示没人叫
	bidCount      int // 叫地主轮数

	// 出牌相关
	currentPlayer     int             // 当前出牌玩家索引
	lastPlayedHand    rule.ParsedHand // 上家出牌
	lastPlayerIdx     int             // 上家索引
	consecutivePasses int             // 连续 PASS 次数

	// 超时控制
	turnTimer        *time.Timer
	offlineWaitTimer *time.Timer   // 离线等待计时器
	remainingTime    time.Duration // 暂停时剩余的时间
	timerStartTime   time.Time     // 计时器开始时间
	timerMu          sync.Mutex

	mu sync.RWMutex
}

// NewGameSession 创建游戏会话
func NewGameSession(room *Room) *GameSession {
	players := make([]*GamePlayer, len(room.PlayerOrder))
	for i, id := range room.PlayerOrder {
		rp := room.Players[id]
		players[i] = &GamePlayer{
			ID:   id,
			Name: rp.Client.Name,
			Seat: i,
		}
	}

	return &GameSession{
		room:          room,
		state:         GameStateInit,
		players:       players,
		highestBidder: -1,
	}
}

// Start 开始游戏
func (gs *GameSession) Start() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 创建并洗牌
	gs.deck = card.NewDeck()
	gs.deck.Shuffle()

	// 发牌
	gs.deal()

	// 进入叫地主阶段
	gs.state = GameStateBidding
	gs.room.State = RoomStateBidding

	// 随机选择第一个叫地主的玩家
	gs.currentBidder = rand.Intn(3)

	// 通知叫地主
	gs.notifyBidTurn()
}

// deal 发牌
func (gs *GameSession) deal() {
	// 每人发 17 张
	for i := 0; i < 17; i++ {
		for j := 0; j < 3; j++ {
			gs.players[j].Hand = append(gs.players[j].Hand, gs.deck[0])
			gs.deck = gs.deck[1:]
		}
	}

	// 剩余 3 张为底牌
	gs.landlordCards = gs.deck

	// 排序手牌
	for _, p := range gs.players {
		sort.Slice(p.Hand, func(i, j int) bool {
			return p.Hand[i].Rank > p.Hand[j].Rank
		})
	}

	// 发送手牌给各玩家（先不显示底牌具体内容）
	for _, p := range gs.players {
		client := gs.room.Players[p.ID].Client
		client.SendMessage(protocol.MustNewMessage(protocol.MsgDealCards, protocol.DealCardsPayload{
			Cards:         protocol.CardsToInfos(p.Hand),
			LandlordCards: make([]protocol.CardInfo, 3), // 暂时不显示
		}))
	}
}

// notifyBidTurn 通知当前玩家叫地主
func (gs *GameSession) notifyBidTurn() {
	player := gs.players[gs.currentBidder]
	timeout := gs.room.server.config.Game.BidTimeout

	// 广播叫地主轮次
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgBidTurn, protocol.BidTurnPayload{
		PlayerID: player.ID,
		Timeout:  timeout,
	}))

	// 设置超时
	gs.startBidTimer()
}

// HandleBid 处理叫地主
func (gs *GameSession) HandleBid(playerID string, bid bool) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.state != GameStateBidding {
		return ErrGameNotStart
	}

	currentPlayer := gs.players[gs.currentBidder]
	if currentPlayer.ID != playerID {
		return ErrNotYourTurn
	}

	// 取消超时计时器
	gs.stopTimer()

	gs.bidCount++

	// 广播叫地主结果
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgBidResult, protocol.BidResultPayload{
		PlayerID:   playerID,
		PlayerName: currentPlayer.Name,
		Bid:        bid,
	}))

	if bid {
		gs.highestBidder = gs.currentBidder
		// 确定地主
		gs.setLandlord(gs.currentBidder)
		return nil
	}

	// 下一个玩家叫地主
	gs.currentBidder = (gs.currentBidder + 1) % 3

	// 如果轮了一圈都没人叫，随机指定地主
	if gs.bidCount >= 3 {
		if gs.highestBidder == -1 {
			gs.highestBidder = rand.Intn(3)
		}
		gs.setLandlord(gs.highestBidder)
		return nil
	}

	// 通知下一个玩家叫地主
	gs.notifyBidTurn()
	return nil
}

// setLandlord 设置地主
func (gs *GameSession) setLandlord(idx int) {
	landlord := gs.players[idx]
	landlord.IsLandlord = true

	// 底牌给地主
	landlord.Hand = append(landlord.Hand, gs.landlordCards...)
	sort.Slice(landlord.Hand, func(i, j int) bool {
		return landlord.Hand[i].Rank > landlord.Hand[j].Rank
	})

	// 更新房间玩家状态
	gs.room.Players[landlord.ID].IsLandlord = true

	// 广播地主信息
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgLandlord, protocol.LandlordPayload{
		PlayerID:      landlord.ID,
		PlayerName:    landlord.Name,
		LandlordCards: protocol.CardsToInfos(gs.landlordCards),
	}))

	// 给地主发送更新后的手牌
	client := gs.room.Players[landlord.ID].Client
	client.SendMessage(protocol.MustNewMessage(protocol.MsgDealCards, protocol.DealCardsPayload{
		Cards:         protocol.CardsToInfos(landlord.Hand),
		LandlordCards: protocol.CardsToInfos(gs.landlordCards),
	}))

	// 开始游戏，地主先出牌
	gs.state = GameStatePlaying
	gs.room.State = RoomStatePlaying
	gs.currentPlayer = idx
	gs.lastPlayerIdx = idx

	gs.notifyPlayTurn()
}

// notifyPlayTurn 通知当前玩家出牌
func (gs *GameSession) notifyPlayTurn() {
	player := gs.players[gs.currentPlayer]
	timeout := gs.room.server.config.Game.TurnTimeout

	// 判断是否必须出牌（新一轮开始）
	mustPlay := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()

	// 判断是否有牌能打过上家
	canBeat := mustPlay || rule.CanBeatWithHand(player.Hand, gs.lastPlayedHand)

	// 广播出牌轮次
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgPlayTurn, protocol.PlayTurnPayload{
		PlayerID: player.ID,
		Timeout:  timeout,
		MustPlay: mustPlay,
		CanBeat:  canBeat,
	}))

	// 设置超时
	gs.startPlayTimer()
}

// HandlePlayCards 处理出牌
func (gs *GameSession) HandlePlayCards(playerID string, cardInfos []protocol.CardInfo) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.state != GameStatePlaying {
		return ErrGameNotStart
	}

	currentPlayer := gs.players[gs.currentPlayer]
	if currentPlayer.ID != playerID {
		return ErrNotYourTurn
	}

	// 取消超时计时器
	gs.stopTimer()

	// 转换牌
	cards := protocol.InfosToCards(cardInfos)

	// 验证牌是否在手中
	if !gs.validateCardsInHand(currentPlayer, cards) {
		return ErrInvalidCards
	}

	// 解析牌型
	handToPlay, err := rule.ParseHand(cards)
	if err != nil {
		return ErrInvalidCards
	}

	// 检查是否能打过上家
	isNewRound := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()
	if !isNewRound && !rule.CanBeat(handToPlay, gs.lastPlayedHand) {
		return ErrCannotBeat
	}

	// 出牌成功，更新状态
	gs.lastPlayedHand = handToPlay
	gs.lastPlayerIdx = gs.currentPlayer
	gs.consecutivePasses = 0

	// 从手牌中移除
	currentPlayer.Hand = card.RemoveCards(currentPlayer.Hand, cards)

	// 广播出牌信息
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgCardPlayed, protocol.CardPlayedPayload{
		PlayerID:   playerID,
		PlayerName: currentPlayer.Name,
		Cards:      cardInfos,
		CardsLeft:  len(currentPlayer.Hand),
		HandType:   handToPlay.Type.String(),
	}))

	// 检查是否获胜
	if len(currentPlayer.Hand) == 0 {
		gs.endGame(currentPlayer)
		return nil
	}

	// 下一个玩家
	gs.currentPlayer = (gs.currentPlayer + 1) % 3
	gs.notifyPlayTurn()

	return nil
}

// HandlePass 处理不出
func (gs *GameSession) HandlePass(playerID string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.state != GameStatePlaying {
		return ErrGameNotStart
	}

	currentPlayer := gs.players[gs.currentPlayer]
	if currentPlayer.ID != playerID {
		return ErrNotYourTurn
	}

	// 检查是否必须出牌
	mustPlay := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()
	if mustPlay {
		return ErrMustPlay
	}

	// 取消超时计时器
	gs.stopTimer()

	gs.consecutivePasses++

	// 广播不出
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgPlayerPass, protocol.PlayerPassPayload{
		PlayerID:   playerID,
		PlayerName: currentPlayer.Name,
	}))

	// 如果连续两人不出，新一轮开始
	if gs.consecutivePasses >= 2 {
		gs.lastPlayedHand = rule.ParsedHand{}
		gs.lastPlayerIdx = (gs.currentPlayer + 1) % 3
		gs.consecutivePasses = 0
	}

	// 下一个玩家
	gs.currentPlayer = (gs.currentPlayer + 1) % 3
	gs.notifyPlayTurn()

	return nil
}

// endGame 结束游戏
func (gs *GameSession) endGame(winner *GamePlayer) {
	gs.state = GameStateEnded
	gs.room.State = RoomStateEnded

	// 收集所有玩家剩余手牌
	playerHands := make([]protocol.PlayerHand, len(gs.players))
	for i, p := range gs.players {
		playerHands[i] = protocol.PlayerHand{
			PlayerID:   p.ID,
			PlayerName: p.Name,
			Cards:      protocol.CardsToInfos(p.Hand),
		}
	}

	// 广播游戏结束
	gs.room.broadcast(protocol.MustNewMessage(protocol.MsgGameOver, protocol.GameOverPayload{
		WinnerID:    winner.ID,
		WinnerName:  winner.Name,
		IsLandlord:  winner.IsLandlord,
		PlayerHands: playerHands,
	}))

	log.Printf("🎮 游戏结束！房间 %s，获胜者: %s (%s)",
		gs.room.Code, winner.Name, ternary(winner.IsLandlord, "地主", "农民"))

	// 记录游戏结果到排行榜
	gs.recordGameResults(winner)
}

// recordGameResults 记录游戏结果到排行榜
func (gs *GameSession) recordGameResults(winner *GamePlayer) {
	ctx := context.Background()
	leaderboard := gs.room.server.leaderboard

	// 计算获胜方
	landlordWins := winner.IsLandlord

	for _, p := range gs.players {
		isWinner := false
		if landlordWins {
			// 地主胜利
			isWinner = p.IsLandlord
		} else {
			// 农民胜利
			isWinner = !p.IsLandlord
		}

		// 获取玩家名称
		playerName := p.Name
		if rp, exists := gs.room.Players[p.ID]; exists && rp.Client != nil {
			playerName = rp.Client.Name
		}

		// 记录结果
		if err := leaderboard.RecordGameResult(ctx, p.ID, playerName, p.IsLandlord, isWinner); err != nil {
			log.Printf("记录游戏结果失败: %v", err)
		}
	}
}

// validateCardsInHand 验证牌是否在手中
func (gs *GameSession) validateCardsInHand(player *GamePlayer, cards []card.Card) bool {
	handCopy := make([]card.Card, len(player.Hand))
	copy(handCopy, player.Hand)

	for _, c := range cards {
		found := false
		for i, h := range handCopy {
			if h.Suit == c.Suit && h.Rank == c.Rank {
				handCopy = append(handCopy[:i], handCopy[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GetPlayerCardsCount 获取玩家手牌数量
func (gs *GameSession) GetPlayerCardsCount(playerID string) int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	for _, p := range gs.players {
		if p.ID == playerID {
			return len(p.Hand)
		}
	}
	return 0
}

// --- 超时控制 ---

func (gs *GameSession) startBidTimer() {
	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	timeout := gs.room.server.config.Game.BidTimeoutDuration()
	gs.timerStartTime = time.Now()
	gs.remainingTime = timeout
	gs.turnTimer = time.AfterFunc(timeout, func() {
		// 超时自动不叫
		currentPlayer := gs.players[gs.currentBidder]
		_ = gs.HandleBid(currentPlayer.ID, false)
	})
}

func (gs *GameSession) startPlayTimer() {
	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	timeout := gs.room.server.config.Game.TurnTimeoutDuration()
	gs.timerStartTime = time.Now()
	gs.remainingTime = timeout
	gs.turnTimer = time.AfterFunc(timeout, func() {
		gs.handlePlayTimeout()
	})
}

func (gs *GameSession) handlePlayTimeout() {
	gs.mu.Lock()

	if gs.state != GameStatePlaying {
		gs.mu.Unlock()
		return
	}

	currentPlayer := gs.players[gs.currentPlayer]

	// 尝试找到最小能打过的牌
	cardsToPlay := rule.FindSmallestBeatingCards(currentPlayer.Hand, gs.lastPlayedHand)

	if cardsToPlay != nil {
		// 找到了能打的牌，出牌
		playerID := currentPlayer.ID
		cardInfos := protocol.CardsToInfos(cardsToPlay)
		gs.mu.Unlock()
		_ = gs.HandlePlayCards(playerID, cardInfos)
		return
	}

	// 没有能打的牌，自动 PASS
	playerID := currentPlayer.ID
	gs.mu.Unlock()
	_ = gs.HandlePass(playerID)
}

func (gs *GameSession) stopTimer() {
	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	if gs.turnTimer != nil {
		gs.turnTimer.Stop()
		gs.turnTimer = nil
	}
	if gs.offlineWaitTimer != nil {
		gs.offlineWaitTimer.Stop()
		gs.offlineWaitTimer = nil
	}
}

// --- 离线处理 ---

// PlayerOffline 玩家离线
func (gs *GameSession) PlayerOffline(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 找到玩家
	playerIdx := -1
	for i, p := range gs.players {
		if p.ID == playerID {
			p.IsOffline = true
			playerIdx = i
			break
		}
	}

	if playerIdx == -1 {
		return
	}

	// 检查是否是当前回合玩家
	isBidding := gs.state == GameStateBidding && gs.currentBidder == playerIdx
	isPlaying := gs.state == GameStatePlaying && gs.currentPlayer == playerIdx

	if !isBidding && !isPlaying {
		return // 不是当前回合，无需暂停
	}

	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	// 暂停计时器，计算剩余时间
	if gs.turnTimer != nil {
		gs.turnTimer.Stop()
		gs.remainingTime = time.Until(gs.timerStartTime.Add(gs.remainingTime))
		if gs.remainingTime < 0 {
			gs.remainingTime = 0
		}
		gs.turnTimer = nil
	}

	// 启动离线等待计时器
	gs.offlineWaitTimer = time.AfterFunc(offlineWaitTimeout, func() {
		gs.handleOfflineTimeout(playerID)
	})

	log.Printf("⏸️ 玩家 %s 离线，暂停计时等待重连 (%v)", gs.players[playerIdx].Name, offlineWaitTimeout)
}

// PlayerOnline 玩家上线
func (gs *GameSession) PlayerOnline(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 找到玩家
	playerIdx := -1
	for i, p := range gs.players {
		if p.ID == playerID {
			p.IsOffline = false
			playerIdx = i
			break
		}
	}

	if playerIdx == -1 {
		return
	}

	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	// 取消离线等待计时器
	if gs.offlineWaitTimer != nil {
		gs.offlineWaitTimer.Stop()
		gs.offlineWaitTimer = nil
	}

	// 检查是否是当前回合玩家，如果是则恢复计时器
	isBidding := gs.state == GameStateBidding && gs.currentBidder == playerIdx
	isPlaying := gs.state == GameStatePlaying && gs.currentPlayer == playerIdx

	if !isBidding && !isPlaying {
		return
	}

	// 恢复计时器
	if gs.remainingTime > 0 {
		gs.timerStartTime = time.Now()
		if isBidding {
			gs.turnTimer = time.AfterFunc(gs.remainingTime, func() {
				currentPlayer := gs.players[gs.currentBidder]
				_ = gs.HandleBid(currentPlayer.ID, false)
			})
		} else {
			gs.turnTimer = time.AfterFunc(gs.remainingTime, func() {
				gs.handlePlayTimeout()
			})
		}
		log.Printf("▶️ 玩家 %s 重连，恢复计时 (剩余 %v)", gs.players[playerIdx].Name, gs.remainingTime)
	}
}

// handleOfflineTimeout 离线超时处理
func (gs *GameSession) handleOfflineTimeout(playerID string) {
	gs.mu.Lock()

	// 找到玩家
	playerIdx := -1
	for i, p := range gs.players {
		if p.ID == playerID {
			playerIdx = i
			break
		}
	}

	if playerIdx == -1 {
		gs.mu.Unlock()
		return
	}

	log.Printf("⏰ 玩家 %s 离线超时，自动执行操作", gs.players[playerIdx].Name)

	// 根据当前状态执行自动操作
	if gs.state == GameStateBidding && gs.currentBidder == playerIdx {
		gs.mu.Unlock()
		_ = gs.HandleBid(playerID, false)
		return
	}

	if gs.state == GameStatePlaying && gs.currentPlayer == playerIdx {
		currentPlayer := gs.players[playerIdx]
		mustPlay := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()

		if mustPlay && len(currentPlayer.Hand) > 0 {
			// 出最小的牌
			minCard := currentPlayer.Hand[len(currentPlayer.Hand)-1]
			gs.mu.Unlock()
			_ = gs.HandlePlayCards(playerID, []protocol.CardInfo{protocol.CardToInfo(minCard)})
			return
		}
		gs.mu.Unlock()
		_ = gs.HandlePass(playerID)
		return
	}

	gs.mu.Unlock()
}

// --- 错误定义 ---

var (
	ErrGameNotStart = &RoomError{Code: protocol.ErrCodeGameNotStart, Message: "游戏尚未开始"}
	ErrNotYourTurn  = &RoomError{Code: protocol.ErrCodeNotYourTurn, Message: "还没轮到您"}
	ErrInvalidCards = &RoomError{Code: protocol.ErrCodeInvalidCards, Message: "无效的牌型"}
	ErrCannotBeat   = &RoomError{Code: protocol.ErrCodeCannotBeat, Message: "您的牌打不过上家"}
	ErrMustPlay     = &RoomError{Code: protocol.ErrCodeMustPlay, Message: "您必须出牌"}
)

// ternary 三元表达式
func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

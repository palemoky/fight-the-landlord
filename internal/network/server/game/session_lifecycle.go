package game

import (
	"context"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

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
	gs.bottomCards = gs.deck

	// 排序手牌
	for _, p := range gs.players {
		sort.Slice(p.Hand, func(i, j int) bool {
			return p.Hand[i].Rank > p.Hand[j].Rank
		})
	}

	// 发送手牌给各玩家（先不显示底牌具体内容）
	for _, p := range gs.players {
		client := gs.room.Players[p.ID].Client
		client.SendMessage(encoding.MustNewMessage(protocol.MsgDealCards, protocol.DealCardsPayload{
			Cards:         convert.CardsToInfos(p.Hand),
			LandlordCards: make([]protocol.CardInfo, 3), // 暂时不显示
		}))
	}
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
			Cards:      convert.CardsToInfos(p.Hand),
		}
	}

	// 广播游戏结束
	gs.room.broadcast(encoding.MustNewMessage(protocol.MsgGameOver, protocol.GameOverPayload{
		WinnerID:    winner.ID,
		WinnerName:  winner.Name,
		IsLandlord:  winner.IsLandlord,
		PlayerHands: playerHands,
	}))

	role := "农民"
	if winner.IsLandlord {
		role = "地主"
	}
	log.Printf("🎮 游戏结束！房间 %s，获胜者: %s (%s)",
		gs.room.Code, winner.Name, role)

	// 记录游戏结果到排行榜
	gs.recordGameResults(winner)

	// 延迟清理房间，让玩家有时间返回大厅查看维护通知
	cleanupDelay := 2 * time.Hour
	log.Printf("⏰ 房间 %s 将在 %v 后自动清理", gs.room.Code, cleanupDelay)

	go func() {
		time.Sleep(cleanupDelay)

		// 让所有玩家离开房间
		gs.room.mu.RLock()
		playerIDs := make([]string, 0, len(gs.room.Players))
		for id := range gs.room.Players {
			playerIDs = append(playerIDs, id)
		}
		gs.room.mu.RUnlock()

		// 逐个让玩家离开房间
		for _, playerID := range playerIDs {
			gs.room.mu.RLock()
			if rp, exists := gs.room.Players[playerID]; exists && rp.Client != nil {
				client := rp.Client
				gs.room.mu.RUnlock()
				gs.room.GetServer().GetRoomManager().(*RoomManager).LeaveRoom(client)
			} else {
				gs.room.mu.RUnlock()
			}
		}

		log.Printf("🧹 房间 %s 已自动清理", gs.room.Code)
	}()
}

// recordGameResults 记录游戏结果到排行榜
func (gs *GameSession) recordGameResults(winner *GamePlayer) {
	ctx := context.Background()
	leaderboard := gs.room.GetServer().GetLeaderboard()

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
			playerName = rp.Client.GetName()
		}

		// 记录结果
		if err := leaderboard.RecordGameResult(ctx, p.ID, playerName, p.IsLandlord, isWinner); err != nil {
			log.Printf("记录游戏结果失败: %v", err)
		}
	}
}

package session

import (
	"context"
	"log"
	"math/rand/v2"
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
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
	gs.room.SetState(RoomStateBidding)

	// 随机选择第一个叫地主的玩家
	gs.currentBidder = rand.IntN(3)

	// 通知叫地主
	gs.notifyBidTurn()
}

// deal 发牌
func (gs *GameSession) deal() {
	// 每人发 17 张
	for range 17 {
		for i := range 3 {
			gs.players[i].Hand = append(gs.players[i].Hand, gs.deck[0])
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
		rp := gs.room.GetPlayer(p.ID)
		client := rp.GetClient()
		client.SendMessage(codec.MustNewMessage(protocol.MsgDealCards, protocol.DealCardsPayload{
			Cards:       convert.CardsToInfos(p.Hand),
			BottomCards: make([]protocol.CardInfo, 3), // 暂时不显示
		}))
	}
}

// endGame 结束游戏
func (gs *GameSession) endGame(winner *GamePlayer) {
	gs.state = GameStateEnded
	gs.room.SetState(RoomStateEnded)

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
	gs.room.Broadcast(codec.MustNewMessage(protocol.MsgGameOver, protocol.GameOverPayload{
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
		gs.room.GetCode(), winner.Name, role)

	// 游戏结束，解散房间
	for _, p := range gs.players {
		rp := gs.room.GetPlayer(p.ID)
		if rp != nil {
			rp.GetClient().SetRoom("")
		}
	}

	// 记录游戏结果到排行榜
	gs.recordGameResults(winner)
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
			isWinner = p.IsLandlord
		} else {
			isWinner = !p.IsLandlord
		}

		// 获取玩家名称
		playerName := p.Name
		rp := gs.room.GetPlayer(p.ID)
		if rp != nil {
			playerName = rp.GetClient().GetName()
		}

		// 记录结果
		if err := leaderboard.RecordGameResult(ctx, p.ID, playerName, p.IsLandlord, isWinner); err != nil {
			log.Printf("记录游戏结果失败: %v", err)
		}
	}
}

package session

import (
	"math/rand"
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/codec"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
)

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
	gs.room.Broadcast(codec.MustNewMessage(protocol.MsgBidResult, protocol.BidResultPayload{
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
	landlord.Hand = append(landlord.Hand, gs.bottomCards...)
	sort.Slice(landlord.Hand, func(i, j int) bool {
		return landlord.Hand[i].Rank > landlord.Hand[j].Rank
	})

	// 更新房间玩家状态
	gs.room.SetPlayerLandlord(landlord.ID)

	// 广播地主信息
	gs.room.Broadcast(codec.MustNewMessage(protocol.MsgLandlord, protocol.LandlordPayload{
		PlayerID:    landlord.ID,
		PlayerName:  landlord.Name,
		BottomCards: convert.CardsToInfos(gs.bottomCards),
	}))

	// 给地主发送更新后的手牌
	rp := gs.room.GetPlayer(landlord.ID)
	client := rp.GetClient()
	client.SendMessage(codec.MustNewMessage(protocol.MsgDealCards, protocol.DealCardsPayload{
		Cards:       convert.CardsToInfos(landlord.Hand),
		BottomCards: convert.CardsToInfos(gs.bottomCards),
	}))

	// 开始游戏，地主先出牌
	gs.state = GameStatePlaying
	gs.room.SetState(RoomStatePlaying)
	gs.currentPlayer = idx
	gs.lastPlayerIdx = idx

	gs.notifyPlayTurn()
}

// notifyBidTurn 通知当前玩家叫地主
func (gs *GameSession) notifyBidTurn() {
	player := gs.players[gs.currentBidder]
	gs.room.Broadcast(codec.MustNewMessage(protocol.MsgBidTurn, protocol.BidTurnPayload{
		PlayerID: player.ID,
		Timeout:  30,
	}))
	gs.startBidTimer()
}

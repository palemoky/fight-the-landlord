package session

import (
	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/server/types"
)

// BuildGameStateDTO 构建游戏状态 DTO（用于重连等场景）
func (gs *GameSession) BuildGameStateDTO(playerID string, sessionManager types.SessionManagerInterface) *protocol.GameStateDTO {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	var hand []card.Card
	for _, p := range gs.players {
		if p.ID == playerID {
			hand = p.Hand
			break
		}
	}
	players := make([]protocol.PlayerInfo, len(gs.players))
	for i, p := range gs.players {
		players[i] = protocol.PlayerInfo{
			ID:         p.ID,
			Name:       p.Name,
			Seat:       p.Seat,
			IsLandlord: p.IsLandlord,
			CardsCount: len(p.Hand),
			Online:     sessionManager.IsOnline(p.ID),
		}
	}
	phase := "waiting"
	switch gs.state {
	case GameStateBidding:
		phase = "bidding"
	case GameStatePlaying:
		phase = "playing"
	case GameStateEnded:
		phase = "ended"
	}
	currentTurnID := ""
	switch gs.state {
	case GameStateBidding:
		currentTurnID = gs.players[gs.currentBidder].ID
	case GameStatePlaying:
		currentTurnID = gs.players[gs.currentPlayer].ID
	}
	var lastPlayed []card.Card
	lastPlayerID := ""
	if !gs.lastPlayedHand.IsEmpty() {
		lastPlayed = gs.lastPlayedHand.Cards
		lastPlayerID = gs.players[gs.lastPlayerIdx].ID
	}
	return &protocol.GameStateDTO{
		Phase:         phase,
		Players:       players,
		Hand:          convert.CardsToInfos(hand),
		LandlordCards: convert.CardsToInfos(gs.bottomCards),
		CurrentTurn:   currentTurnID,
		LastPlayed:    convert.CardsToInfos(lastPlayed),
		LastPlayerID:  lastPlayerID,
		MustPlay:      gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty(),
		CanBeat:       true,
	}
}

// SerializeForRedis 为Redis序列化准备数据（提供只读访问）
func (gs *GameSession) SerializeForRedis(serialize func()) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	serialize()
}

// GetStateForSerialization 获取state用于序列化
func (gs *GameSession) GetStateForSerialization() GameState {
	return gs.state
}

// GetCurrentPlayerForSerialization 获取currentPlayer用于序列化
func (gs *GameSession) GetCurrentPlayerForSerialization() int {
	return gs.currentPlayer
}

// GetHighestBidderForSerialization 获取highestBidder用于序列化
func (gs *GameSession) GetHighestBidderForSerialization() int {
	return gs.highestBidder
}

// GetPlayersForSerialization 获取players用于序列化
func (gs *GameSession) GetPlayersForSerialization() []*GamePlayer {
	return gs.players
}

// GetBottomCardsForSerialization 获取bottomCards用于序列化
func (gs *GameSession) GetBottomCardsForSerialization() []card.Card {
	return gs.bottomCards
}

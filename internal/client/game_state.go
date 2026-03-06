package client

import (
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

// GameState 管理客户端侧的游戏状态
type GameState struct {
	// 玩家数据
	Hand        []card.Card
	BottomCards []card.Card
	IsLandlord  bool

	// 其他玩家
	Players []protocol.PlayerInfo

	// 游戏进程
	RoomCode       string
	CurrentTurn    string
	LastPlayedBy   string
	LastPlayedName string
	LastPlayed     []card.Card
	LastHandType   string

	// 游戏结果
	Winner           string
	WinnerIsLandlord bool

	// 功能组件
	CardCounter *CardCounter
}

// NewGameState 创建一个新的游戏状态
func NewGameState() *GameState {
	return &GameState{
		CardCounter: NewCardCounter(),
	}
}

// SortHand 将玩家手牌按点数降序排序
func (gs *GameState) SortHand() {
	sort.Slice(gs.Hand, func(i, j int) bool {
		return gs.Hand[i].Rank > gs.Hand[j].Rank
	})
}

// Reset 清除所有游戏状态
func (gs *GameState) Reset() {
	gs.Hand = nil
	gs.BottomCards = nil
	gs.Players = nil
	gs.RoomCode = ""
	gs.CurrentTurn = ""
	gs.LastPlayedBy = ""
	gs.LastPlayedName = ""
	gs.LastPlayed = nil
	gs.LastHandType = ""
	gs.IsLandlord = false
	gs.Winner = ""
	gs.WinnerIsLandlord = false
	gs.CardCounter = NewCardCounter()
}

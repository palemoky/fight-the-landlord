package client

import (
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// GameState manages client-side game state
type GameState struct {
	// Player data
	Hand        []card.Card
	BottomCards []card.Card
	IsLandlord  bool

	// Other players
	Players []protocol.PlayerInfo

	// Game progress
	RoomCode       string
	CurrentTurn    string
	LastPlayedBy   string
	LastPlayedName string
	LastPlayed     []card.Card
	LastHandType   string

	// Game result
	Winner           string
	WinnerIsLandlord bool

	// Features
	CardCounter *CardCounter
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	return &GameState{
		CardCounter: NewCardCounter(),
	}
}

// SortHand sorts the player's hand in descending order by rank
func (gs *GameState) SortHand() {
	sort.Slice(gs.Hand, func(i, j int) bool {
		return gs.Hand[i].Rank > gs.Hand[j].Rank
	})
}

// Reset clears all game state
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

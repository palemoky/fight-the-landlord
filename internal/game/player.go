package game

import (
	"sort"

	"github.com/palemoky/fight-the-landlord-go/internal/card"
)

// Player 定义玩家
type Player struct {
	Name       string
	Hand       []card.Card 
	IsLandlord bool
}

func (p *Player) SortHand() {
	sort.Slice(p.Hand, func(i, j int) bool {
		return p.Hand[i].Rank < p.Hand[j].Rank
	})
}
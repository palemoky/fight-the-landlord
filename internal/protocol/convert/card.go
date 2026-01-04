package convert

import (
	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
)

// CardToInfo 将 card.Card 转换为 protocol.CardInfo
func CardToInfo(c card.Card) protocol.CardInfo {
	return protocol.CardInfo{
		Suit:  int(c.Suit),
		Rank:  int(c.Rank),
		Color: int(c.Color),
	}
}

// CardsToInfos 将 []card.Card 转换为 []protocol.CardInfo
func CardsToInfos(cards []card.Card) []protocol.CardInfo {
	infos := make([]protocol.CardInfo, len(cards))
	for i, c := range cards {
		infos[i] = CardToInfo(c)
	}
	return infos
}

// InfoToCard 将 protocol.CardInfo 转换为 card.Card
func InfoToCard(info protocol.CardInfo) card.Card {
	return card.Card{
		Suit:  card.Suit(info.Suit),
		Rank:  card.Rank(info.Rank),
		Color: card.CardColor(info.Color),
	}
}

// InfosToCards 将 []protocol.CardInfo 转换为 []card.Card
func InfosToCards(infos []protocol.CardInfo) []card.Card {
	cards := make([]card.Card, len(infos))
	for i, info := range infos {
		cards[i] = InfoToCard(info)
	}
	return cards
}

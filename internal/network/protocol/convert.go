package protocol

import "github.com/palemoky/fight-the-landlord-go/internal/card"

// CardToInfo 将 card.Card 转换为 CardInfo
func CardToInfo(c card.Card) CardInfo {
	return CardInfo{
		Suit:  int(c.Suit),
		Rank:  int(c.Rank),
		Color: int(c.Color),
	}
}

// CardsToInfos 将 []card.Card 转换为 []CardInfo
func CardsToInfos(cards []card.Card) []CardInfo {
	infos := make([]CardInfo, len(cards))
	for i, c := range cards {
		infos[i] = CardToInfo(c)
	}
	return infos
}

// InfoToCard 将 CardInfo 转换为 card.Card
func InfoToCard(info CardInfo) card.Card {
	return card.Card{
		Suit:  card.Suit(info.Suit),
		Rank:  card.Rank(info.Rank),
		Color: card.CardColor(info.Color),
	}
}

// InfosToCards 将 []CardInfo 转换为 []card.Card
func InfosToCards(infos []CardInfo) []card.Card {
	cards := make([]card.Card, len(infos))
	for i, info := range infos {
		cards[i] = InfoToCard(info)
	}
	return cards
}

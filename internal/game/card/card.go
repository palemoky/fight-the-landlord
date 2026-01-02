package card

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

// Suit 定义花色
type Suit int

// Rank 定义点数
type Rank int

// CardColor 定义牌的颜色
type CardColor int

const (
	Black CardColor = iota
	Red
)

// Card 定义一张牌
type Card struct {
	Suit  Suit
	Rank  Rank
	Color CardColor
}

const (
	Spade   Suit = iota // 黑桃
	Heart               // 红心
	Club                // 梅花
	Diamond             // 方块
	Joker               // 王牌
)

func (s Suit) String() string {
	switch s {
	case Spade:
		return "♠"
	case Heart:
		return "♥"
	case Club:
		return "♣"
	case Diamond:
		return "♦"
	case Joker:
		return ""
	default:
		return ""
	}
}

const (
	Rank3 Rank = iota + 3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJ // Jack
	RankQ // Queen
	RankK // King
	RankA // Ace
	Rank2
	RankBlackJoker // BlackJoker
	RankRedJoker   // RedJoker
)

func (r Rank) String() string {
	switch r {
	case RankJ:
		return "J"
	case RankQ:
		return "Q"
	case RankK:
		return "K"
	case RankA:
		return "A"
	case Rank2:
		return "2"
	case RankBlackJoker:
		return "B"
	case RankRedJoker:
		return "R"
	default:
		if r == Rank10 {
			return "10"
		}
		return strconv.Itoa(int(r))
	}
}

func RankFromChar(char rune) (Rank, error) {
	switch char {
	case '3':
		return Rank3, nil
	case '4':
		return Rank4, nil
	case '5':
		return Rank5, nil
	case '6':
		return Rank6, nil
	case '7':
		return Rank7, nil
	case '8':
		return Rank8, nil
	case '9':
		return Rank9, nil
	case 'T':
		return Rank10, nil
	case 'J':
		return RankJ, nil
	case 'Q':
		return RankQ, nil
	case 'K':
		return RankK, nil
	case 'A':
		return RankA, nil
	case '2':
		return Rank2, nil
	case 'B':
		return RankBlackJoker, nil
	case 'R':
		return RankRedJoker, nil
	default:
		return -1, fmt.Errorf("无法识别的点数: %c", char)
	}
}

// Deck 定义一副牌
type Deck []Card

func NewDeck() Deck {
	deck := make(Deck, 0, 54)
	for s := Spade; s <= Diamond; s++ {
		for r := Rank3; r <= Rank2; r++ {
			color := Black
			if s == Heart || s == Diamond {
				color = Red
			}
			deck = append(deck, Card{Suit: s, Rank: r, Color: color})
		}
	}
	deck = append(deck, Card{Suit: Joker, Rank: RankBlackJoker, Color: Black})
	deck = append(deck, Card{Suit: Joker, Rank: RankRedJoker, Color: Red})
	return deck
}

func (d Deck) Shuffle() {
	rand.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
}

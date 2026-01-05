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

// suitSymbols 花色符号映射表
var suitSymbols = map[Suit]string{
	Spade:   "♠",
	Heart:   "♥",
	Club:    "♣",
	Diamond: "♦",
	Joker:   "",
}

func (s Suit) String() string {
	if symbol, ok := suitSymbols[s]; ok {
		return symbol
	}
	return ""
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

// rankNames 牌面值字符串映射表
var rankNames = map[Rank]string{
	Rank3:          "3",
	Rank4:          "4",
	Rank5:          "5",
	Rank6:          "6",
	Rank7:          "7",
	Rank8:          "8",
	Rank9:          "9",
	Rank10:         "10",
	RankJ:          "J",
	RankQ:          "Q",
	RankK:          "K",
	RankA:          "A",
	Rank2:          "2",
	RankBlackJoker: "B",
	RankRedJoker:   "R",
}

func (r Rank) String() string {
	if name, ok := rankNames[r]; ok {
		return name
	}
	return strconv.Itoa(int(r))
}

// charToRank 用于快速查找字符对应的 Rank
var charToRank = map[rune]Rank{
	'3': Rank3,
	'4': Rank4,
	'5': Rank5,
	'6': Rank6,
	'7': Rank7,
	'8': Rank8,
	'9': Rank9,
	'T': Rank10,
	'J': RankJ,
	'Q': RankQ,
	'K': RankK,
	'A': RankA,
	'2': Rank2,
	'B': RankBlackJoker,
	'R': RankRedJoker,
}

func RankFromChar(char rune) (Rank, error) {
	if rank, ok := charToRank[char]; ok {
		return rank, nil
	}
	return -1, fmt.Errorf("无法识别的点数: %c", char)
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
	deck = append(deck,
		Card{Suit: Joker, Rank: RankBlackJoker, Color: Black},
		Card{Suit: Joker, Rank: RankRedJoker, Color: Red},
	)
	return deck
}

func (d Deck) Shuffle() {
	rand.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
}

package card

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/pterm/pterm"
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

func (c Card) String() string {
	var cardStr string
	if c.Suit == Joker {
		cardStr = fmt.Sprintf(" %s ", c.Rank.String())
	} else {
		cardStr = fmt.Sprintf("%s%-2s", c.Suit.String(), c.Rank.String())
	}

	if c.Color == Red {
		// 红牌使用红色前景和白色背景
		return pterm.NewStyle(pterm.FgRed, pterm.BgWhite).Sprint(cardStr)
	}
	// 黑牌使用默认颜色（黑色前景）和白色背景
	return pterm.NewStyle(pterm.FgBlack, pterm.BgWhite).Sprint(cardStr)
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
		return "♔"
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
		// 确保 10 不会破坏对齐
		if r == Rank10 {
			return "10"
		}
		return strconv.Itoa(int(r))
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
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
}

func (c Card) RenderCard(str string) string {
	// 手牌的样式
	styleRed := pterm.NewRGBStyle(pterm.NewRGB(192, 0, 0), pterm.NewRGB(210, 196, 191))
	styleBlack := pterm.NewStyle(pterm.FgBlack, pterm.BgWhite)

	// 格式化牌的内容，例如 "♥K  " 或 "♠10 " (左对齐，总宽度4)
	s := fmt.Sprintf("%-2s", str)

	// 根据牌的颜色选择样式
	var rankStyledContent string
	rankStyledContent = styleBlack.Sprint(s)
	if c.Color == Red {
		rankStyledContent = styleRed.Sprint(s)
	}

	return rankStyledContent
}

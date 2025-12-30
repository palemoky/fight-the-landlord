package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
)

func TestFindSmallestBeatingCards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		playerHand   []card.Card
		opponentHand []card.Card
		expected     []card.Card
	}{
		{
			name: "Single: Beat 3 with 4",
			playerHand: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank5, Suit: card.Heart},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank3, Suit: card.Diamond},
			},
			expected: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
			},
		},
		{
			name: "Single: Cannot beat 2 with Ace",
			playerHand: []card.Card{
				{Rank: card.RankA, Suit: card.Spade},
				{Rank: card.RankK, Suit: card.Heart},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank2, Suit: card.Diamond},
			},
			expected: nil,
		},
		{
			name: "Pair: Beat 3s with 4s",
			playerHand: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank5, Suit: card.Club},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank3, Suit: card.Diamond},
				{Rank: card.Rank3, Suit: card.Club},
			},
			expected: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
			},
		},
		{
			name: "Trio: Beat 3s with 4s",
			playerHand: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank3, Suit: card.Diamond},
				{Rank: card.Rank3, Suit: card.Club},
				{Rank: card.Rank3, Suit: card.Heart},
			},
			expected: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
			},
		},
		{
			name: "TrioWithSingle: Beat 333+5 with 444+6",
			playerHand: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank3, Suit: card.Diamond},
				{Rank: card.Rank3, Suit: card.Club},
				{Rank: card.Rank3, Suit: card.Heart},
				{Rank: card.Rank5, Suit: card.Spade},
			},
			expected: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
			},
		},
		{
			name: "TrioWithPair: Beat 333+55 with 444+66",
			playerHand: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
				{Rank: card.Rank6, Suit: card.Spade},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank3, Suit: card.Diamond},
				{Rank: card.Rank3, Suit: card.Club},
				{Rank: card.Rank3, Suit: card.Heart},
				{Rank: card.Rank5, Suit: card.Spade},
				{Rank: card.Rank5, Suit: card.Heart},
			},
			expected: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
				{Rank: card.Rank6, Suit: card.Spade},
			},
		},
		{
			name: "TrioWithPair: Cannot beat without pair kicker",
			playerHand: []card.Card{
				{Rank: card.Rank4, Suit: card.Spade},
				{Rank: card.Rank4, Suit: card.Heart},
				{Rank: card.Rank4, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond}, // single, not pair
			},
			opponentHand: []card.Card{
				{Rank: card.Rank3, Suit: card.Diamond},
				{Rank: card.Rank3, Suit: card.Club},
				{Rank: card.Rank3, Suit: card.Heart},
				{Rank: card.Rank5, Suit: card.Spade},
				{Rank: card.Rank5, Suit: card.Heart},
			},
			expected: nil, // no pair kicker available
		},
		{
			name: "Pair: Beat with Bomb (fallback)",
			playerHand: []card.Card{
				{Rank: card.Rank5, Suit: card.Spade}, // Single 5
				{Rank: card.Rank6, Suit: card.Spade}, // Bomb 6s
				{Rank: card.Rank6, Suit: card.Heart},
				{Rank: card.Rank6, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank2, Suit: card.Diamond},
				{Rank: card.Rank2, Suit: card.Club},
			},
			expected: []card.Card{
				{Rank: card.Rank6, Suit: card.Spade},
				{Rank: card.Rank6, Suit: card.Heart},
				{Rank: card.Rank6, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
			},
		},
		{
			name: "Bomb: Beat smaller bomb with larger bomb",
			playerHand: []card.Card{
				{Rank: card.Rank6, Suit: card.Spade},
				{Rank: card.Rank6, Suit: card.Heart},
				{Rank: card.Rank6, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank5, Suit: card.Spade},
				{Rank: card.Rank5, Suit: card.Heart},
				{Rank: card.Rank5, Suit: card.Club},
				{Rank: card.Rank5, Suit: card.Diamond},
			},
			expected: []card.Card{
				{Rank: card.Rank6, Suit: card.Spade},
				{Rank: card.Rank6, Suit: card.Heart},
				{Rank: card.Rank6, Suit: card.Club},
				{Rank: card.Rank6, Suit: card.Diamond},
			},
		},
		{
			name: "Bomb: Beat bomb with Rocket",
			playerHand: []card.Card{
				{Rank: card.RankBlackJoker, Suit: card.Joker},
				{Rank: card.RankRedJoker, Suit: card.Joker},
				{Rank: card.Rank3, Suit: card.Spade},
			},
			opponentHand: []card.Card{
				{Rank: card.Rank2, Suit: card.Spade},
				{Rank: card.Rank2, Suit: card.Heart},
				{Rank: card.Rank2, Suit: card.Club},
				{Rank: card.Rank2, Suit: card.Diamond},
			},
			expected: []card.Card{
				{Rank: card.RankBlackJoker, Suit: card.Joker},
				{Rank: card.RankRedJoker, Suit: card.Joker},
			},
		},
		{
			name: "New Round: Play smallest single",
			playerHand: []card.Card{
				{Rank: card.RankA, Suit: card.Heart}, // Large
				{Rank: card.Rank5, Suit: card.Spade}, // Smallest (at end)
			},
			opponentHand: nil, // Empty means new round/start
			expected: []card.Card{
				{Rank: card.Rank5, Suit: card.Spade},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Pre-parse opponent hand to simulate game state
			var parsedOpponent ParsedHand
			if tt.opponentHand == nil {
				parsedOpponent = ParsedHand{} // Empty
			} else {
				parsedOpponent, _ = ParseHand(tt.opponentHand)
			}

			result := FindSmallestBeatingCards(tt.playerHand, parsedOpponent)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.expected), len(result))
				// Compare ranks mainly to carry correctness
				for i := range result {
					assert.Equal(t, tt.expected[i].Rank, result[i].Rank)
				}
			}
		})
	}
}

func TestHandType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		handType HandType
		expected string
	}{
		{Invalid, "无效"},
		{Single, "单张"},
		{Pair, "对子"},
		{Trio, "三张"},
		{TrioWithSingle, "三带一"},
		{TrioWithPair, "三带二"},
		{Straight, "顺子"},
		{PairStraight, "连对"},
		{Plane, "飞机"},
		{PlaneWithSingles, "飞机带单"},
		{PlaneWithPairs, "飞机带对"},
		{Bomb, "炸弹"},
		{Rocket, "王炸"},
		{FourWithTwo, "四带二"},
		{FourWithTwoPairs, "四带两对"},
		{HandType(99), "无效"}, // Unknown type returns default
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.handType.String())
		})
	}
}

package card

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDeck 验证 NewDeck 是否能正确创建一副完整的、无重复的54张牌
func TestNewDeck(t *testing.T) {
	assert := assert.New(t)

	// 创建一副新牌
	deck := NewDeck()

	// Assertion 1: 牌的总数必须是54张
	assert.Len(deck, 54, "A new deck should have exactly 54 cards")

	// Assertion 2: 验证牌的构成和唯一性
	rankCounts := make(map[Rank]int)
	colorCounts := make(map[CardColor]int)
	uniquenessCheck := make(map[Card]bool)

	for _, card := range deck {
		rankCounts[card.Rank]++
		colorCounts[card.Color]++
		uniquenessCheck[card] = true
	}

	// 2a: 检查唯一性
	assert.Len(uniquenessCheck, 54, "All cards in the deck must be unique")

	// 2b: 检查点数数量 (3-2各有4张，大小王各1张)
	for r := Rank3; r <= Rank2; r++ {
		assert.Equal(4, rankCounts[r], fmt.Sprintf("Should have 4 cards of rank %s", r.String()))
	}
	assert.Equal(1, rankCounts[RankBlackJoker], "Should have 1 Black Joker")
	assert.Equal(1, rankCounts[RankRedJoker], "Should have 1 Red Joker")

	// 2c: 检查颜色数量 (红黑各26张 + 红王/黑王)
	assert.Equal(27, colorCounts[Red], "Should have 27 Red cards (26 + Red Joker)")
	assert.Equal(27, colorCounts[Black], "Should have 27 Black cards (26 + Black Joker)")
}

// TestDeck_Shuffle 验证洗牌功能
func TestDeck_Shuffle(t *testing.T) {
	require := require.New(t) // require 在失败时会停止测试，适合前置条件
	assert := assert.New(t)

	// Arrange: 创建两副完全相同的牌
	deck1 := NewDeck()
	deck2 := make(Deck, len(deck1))
	copy(deck2, deck1)

	// Pre-condition check
	require.Equal(deck1, deck2, "Copied deck should be identical before shuffle")

	// Action: 洗牌
	deck1.Shuffle()

	// Assertion 1: 洗牌后牌的数量不变
	assert.Len(deck1, 54, "Shuffled deck must still have 54 cards")

	// Assertion 2: 洗牌后顺序应该发生改变
	// 注意: 这是一个概率性测试，极小概率下洗牌后顺序不变，但对于54张牌来说概率可以忽略不计
	assert.NotEqual(deck1, deck2, "Shuffled deck should not be in the same order as the original")

	// Assertion 3: 洗牌后，牌的集合应该和原来完全一样（只是顺序不同）
	assert.ElementsMatch(deck2, deck1, "Shuffled deck should contain the exact same cards as the original")
}

// TestStringers 使用表驱动测试来验证所有 String() 方法的输出
func TestStringers(t *testing.T) {
	// --- 测试 Suit.String() ---
	t.Run("Suit Stringer", func(t *testing.T) {
		suitTests := []struct {
			suit Suit
			want string
		}{
			{Spade, "♠"},
			{Heart, "♥"},
			{Club, "♣"},
			{Diamond, "♦"},
			{Joker, ""},
			{Suit(99), ""}, // Edge case: invalid suit
		}

		for _, tt := range suitTests {
			assert.Equal(t, tt.want, tt.suit.String())
		}
	})

	// --- 测试 Rank.String() ---
	t.Run("Rank Stringer", func(t *testing.T) {
		rankTests := []struct {
			rank Rank
			want string
		}{
			{Rank3, "3"},
			{Rank10, "10"},
			{RankK, "K"},
			{RankA, "A"},
			{Rank2, "2"},
			{RankBlackJoker, "B"},
			{RankRedJoker, "R"},
		}
		for _, tt := range rankTests {
			assert.Equal(t, tt.want, tt.rank.String())
		}
	})
}

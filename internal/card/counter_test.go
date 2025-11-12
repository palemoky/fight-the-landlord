package card

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCardCounter 验证记牌器的初始化是否正确
func TestNewCardCounter(t *testing.T) {
	counter := NewCardCounter()

	// 1. 确保记牌器和其内部map都已成功创建
	require.NotNil(t, counter, "NewCardCounter should not return nil.")
	require.NotNil(t, counter.remainingCards, "The internal map of remaining cards should be initialized.")

	// 2. 验证标准点数牌的数量
	// 我们可以抽样检查几个，比如 K 和 3
	assert.Equal(t, 4, counter.remainingCards[RankK], "There should be 4 Kings initially.")
	assert.Equal(t, 4, counter.remainingCards[Rank3], "There should be 4 Threes initially.")
	assert.Equal(t, 4, counter.remainingCards[Rank2], "There should be 4 Twos initially.")

	// 3. 验证大小王牌的数量
	assert.Equal(t, 1, counter.remainingCards[RankBlackJoker], "There should be 1 Black Joker initially.")
	assert.Equal(t, 1, counter.remainingCards[RankRedJoker], "There should be 1 Red Joker initially.")

	// 4. 确保所有点数都被初始化
	assert.Len(t, counter.remainingCards, 15, "The counter should track all 15 ranks.")
}

// TestCardCounter_Update 测试更新记牌器的逻辑
func TestCardCounter_Update(t *testing.T) {
	// 为了方便，创建一个辅助函数来快速生成牌
	newCard := func(rank Rank) Card {
		// 花色和颜色对于这个测试不重要
		return Card{Suit: Spade, Rank: rank, Color: Black}
	}

	t.Run("Update with a single card", func(t *testing.T) {
		counter := NewCardCounter()
		playedCards := []Card{newCard(RankA)}

		counter.Update(playedCards)

		assert.Equal(t, 3, counter.remainingCards[RankA], "After playing one Ace, there should be 3 left.")
		assert.Equal(t, 4, counter.remainingCards[RankK], "Playing an Ace should not affect the count of Kings.")
	})

	t.Run("Update with multiple cards of the same rank", func(t *testing.T) {
		counter := NewCardCounter()
		playedCards := []Card{newCard(Rank7), newCard(Rank7), newCard(Rank7)}

		counter.Update(playedCards)

		assert.Equal(t, 1, counter.remainingCards[Rank7], "After playing three 7s, there should be 1 left.")
	})

	t.Run("Update with a complex hand (Trio with single)", func(t *testing.T) {
		counter := NewCardCounter()
		playedCards := []Card{
			newCard(RankJ), newCard(RankJ), newCard(RankJ), // Trio of Jacks
			newCard(Rank5), // Single 5
		}

		counter.Update(playedCards)

		assert.Equal(t, 1, counter.remainingCards[RankJ], "After playing a trio of Jacks, there should be 1 left.")
		assert.Equal(t, 3, counter.remainingCards[Rank5], "After playing a single 5, there should be 3 left.")
	})

	t.Run("Update with a Joker", func(t *testing.T) {
		counter := NewCardCounter()
		playedCards := []Card{newCard(RankRedJoker)}

		counter.Update(playedCards)

		assert.Equal(t, 0, counter.remainingCards[RankRedJoker], "After playing the Red Joker, there should be 0 left.")
		assert.Equal(t, 1, counter.remainingCards[RankBlackJoker], "Playing the Red Joker should not affect the Black Joker.")
	})

	t.Run("Update with an empty slice", func(t *testing.T) {
		counter := NewCardCounter()
		initialState := counter.GetRemainingCards()

		// 复制一份初始状态用于比较
		expectedState := make(map[Rank]int)
		maps.Copy(expectedState, initialState)

		counter.Update([]Card{})

		assert.Equal(t, expectedState, counter.remainingCards, "Updating with an empty slice should not change the counter's state.")
	})
}

// TestCardCounter_GetRemainingCards 测试获取剩余牌的功能
func TestCardCounter_GetRemainingCards(t *testing.T) {
	counter := NewCardCounter()
	remaining := counter.GetRemainingCards()

	// 1. 确保返回的 map 非空
	require.NotNil(t, remaining, "GetRemainingCards should not return a nil map.")

	// 2. 确保返回的 map 内容与内部状态一致
	assert.Equal(t, counter.remainingCards, remaining, "The returned map should be equal to the internal state.")

	// 3. (可选) 验证它返回的是引用而不是拷贝
	// 这有助于了解该函数的行为
	t.Run("Returns a reference, not a copy", func(t *testing.T) {
		c := NewCardCounter()
		rem := c.GetRemainingCards()

		// 修改返回的 map
		rem[RankK] = 99

		// 检查内部状态是否也被修改
		assert.Equal(t, 99, c.remainingCards[RankK], "Modifying the returned map should also modify the internal state, as it's a reference.")
	})
}

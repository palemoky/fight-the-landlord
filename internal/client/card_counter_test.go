package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
)

func TestNewCardCounter(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()

	require.NotNil(t, cc, "NewCardCounter should not return nil")

	remaining := cc.GetRemaining()

	// Test that all ranks from 3 to A have 4 cards
	for rank := card.Rank3; rank <= card.RankA; rank++ {
		assert.Equal(t, 4, remaining[rank], "Rank %v should have 4 cards", rank)
	}

	// Test Rank2 has 4 cards
	assert.Equal(t, 4, remaining[card.Rank2], "Rank2 should have 4 cards")

	// Test Jokers have 1 card each
	assert.Equal(t, 1, remaining[card.RankBlackJoker], "BlackJoker should have 1 card")
	assert.Equal(t, 1, remaining[card.RankRedJoker], "RedJoker should have 1 card")

	// Total should be 54 cards
	total := 0
	for _, count := range remaining {
		total += count
	}
	assert.Equal(t, 54, total, "Total cards should be 54")
}

func TestCardCounter_Reset(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()

	// Deduct some cards
	testCards := []card.Card{
		{Rank: card.Rank3},
		{Rank: card.Rank3},
		{Rank: card.RankA},
	}
	cc.DeductCards(testCards)

	// Verify cards were deducted
	assert.Equal(t, 2, cc.GetRemaining()[card.Rank3], "After deducting 2 Rank3, should have 2 left")

	// Reset
	cc.Reset()

	// Verify all cards are back to original state
	remaining := cc.GetRemaining()
	assert.Equal(t, 4, remaining[card.Rank3], "After reset, Rank3 should have 4 cards")
	assert.Equal(t, 4, remaining[card.RankA], "After reset, RankA should have 4 cards")
}

func TestCardCounter_DeductCards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cardsToDeduct []card.Card
		expectedRank3 int
		expectedRankA int
	}{
		{
			name: "deduct single card",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
			},
			expectedRank3: 3,
			expectedRankA: 4,
		},
		{
			name: "deduct multiple same rank",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.Rank3},
				{Rank: card.Rank3},
			},
			expectedRank3: 1,
			expectedRankA: 4,
		},
		{
			name: "deduct all of one rank",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.Rank3},
				{Rank: card.Rank3},
				{Rank: card.Rank3},
			},
			expectedRank3: 0,
			expectedRankA: 4,
		},
		{
			name: "deduct mixed ranks",
			cardsToDeduct: []card.Card{
				{Rank: card.Rank3},
				{Rank: card.RankA},
				{Rank: card.Rank3},
			},
			expectedRank3: 2,
			expectedRankA: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cc := NewCardCounter()
			cc.DeductCards(tt.cardsToDeduct)

			remaining := cc.GetRemaining()
			assert.Equal(t, tt.expectedRank3, remaining[card.Rank3], "Rank3 count mismatch")
			assert.Equal(t, tt.expectedRankA, remaining[card.RankA], "RankA count mismatch")
		})
	}
}

func TestCardCounter_DeductCards_OverDeduct(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()

	// Try to deduct more cards than available
	manyCards := make([]card.Card, 10)
	for i := range manyCards {
		manyCards[i] = card.Card{Rank: card.Rank3}
	}

	cc.DeductCards(manyCards)

	// Should not go below 0
	assert.GreaterOrEqual(t, cc.GetRemaining()[card.Rank3], 0, "Card count should not go below 0")
}

func TestCardCounter_DeductCards_EmptySlice(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()
	initialRemaining := cc.GetRemaining()

	// Deduct empty slice
	cc.DeductCards([]card.Card{})

	// Nothing should change
	afterRemaining := cc.GetRemaining()
	for rank := card.Rank3; rank <= card.RankRedJoker; rank++ {
		assert.Equal(t, initialRemaining[rank], afterRemaining[rank],
			"Rank %v count should not change after deducting empty slice", rank)
	}
}

func TestCardCounter_GetRemaining(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()
	remaining := cc.GetRemaining()

	require.NotNil(t, remaining, "GetRemaining should not return nil")

	// Verify it returns the actual map (not a copy)
	// Modifying the returned map should affect the counter
	originalCount := remaining[card.Rank3]
	remaining[card.Rank3] = 99

	assert.Equal(t, 99, cc.GetRemaining()[card.Rank3],
		"GetRemaining should return the actual map, not a copy")

	// Restore for other tests
	remaining[card.Rank3] = originalCount
}

// --- 游戏场景测试 ---

// countTotalCards 计算记牌器中剩余牌的总数
func countTotalCards(cc *CardCounter) int {
	total := 0
	for _, count := range cc.GetRemaining() {
		total += count
	}
	return total
}

func TestCardCounter_GameScenario_InitialDeck(t *testing.T) {
	t.Parallel()

	// 发牌前共54张牌
	cc := NewCardCounter()
	assert.Equal(t, 54, countTotalCards(cc), "一副牌共54张")
}

func TestCardCounter_GameScenario_AfterDeal(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()
	deck := card.NewDeck() // 使用真实的一副牌
	// 为了测试稳定，这里我们不洗牌，直接取前17张
	playerHand := deck[:17]

	cc.DeductCards(playerHand)
	assert.Equal(t, 37, countTotalCards(cc), "发牌后扣除自己17张牌，记牌器应剩余37张")
}

func TestCardCounter_GameScenario_LandlordVsFarmer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		isLandlord    bool
		expectedCards int
	}{
		{
			name:          "地主视角：37-3=34张",
			isLandlord:    true,
			expectedCards: 34,
		},
		{
			name:          "农民视角：保持37张",
			isLandlord:    false,
			expectedCards: 37,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cc := NewCardCounter()
			deck := card.NewDeck()

			// 模拟玩家手牌（17张）
			playerHand := deck[:17]
			cc.DeductCards(playerHand)

			// 底牌（3张）
			bottomCards := deck[17:20]

			// 地主需要扣除底牌
			if tc.isLandlord {
				cc.DeductCards(bottomCards)
			}

			assert.Equal(t, tc.expectedCards, countTotalCards(cc))
		})
	}
}

// 每个玩家的记牌器 = 另外两个玩家的牌 + 底牌
func TestCardCounter_GameScenario_CounterEqualsOtherPlayers(t *testing.T) {
	t.Parallel()

	deck := card.NewDeck()
	// 分配牌
	player1Hand := deck[0:17]
	player2Hand := deck[17:34]
	player3Hand := deck[34:51]
	bottomCards := deck[51:54]

	// 验证总数
	totalCards := len(player1Hand) + len(player2Hand) + len(player3Hand) + len(bottomCards)
	assert.Equal(t, 54, totalCards, "总牌数应为54张")

	// 玩家1的记牌器
	cc1 := NewCardCounter()
	// 扣除玩家1自己的牌
	cc1.DeductCards(player1Hand)

	// 记牌器应该等于其他玩家的牌 + 底牌
	expectedInCounter := len(player2Hand) + len(player3Hand) + len(bottomCards)
	assert.Equal(t, expectedInCounter, countTotalCards(cc1),
		"玩家记牌器 = 其他两人牌数 + 底牌 = %d", expectedInCounter)

	// 将剩余牌统计，验证具体牌的分布
	expectedMap := make(map[card.Rank]int)
	for _, c := range player2Hand {
		expectedMap[c.Rank]++
	}
	for _, c := range player3Hand {
		expectedMap[c.Rank]++
	}
	for _, c := range bottomCards {
		expectedMap[c.Rank]++
	}

	remaining := cc1.GetRemaining()
	for rank, count := range expectedMap {
		assert.Equal(t, count, remaining[rank], "Rank %v 数量不匹配", rank)
	}
}

func TestCardCounter_GameScenario_DuringPlay(t *testing.T) {
	t.Parallel()

	cc := NewCardCounter()
	deck := card.NewDeck()

	// 玩家1手牌
	playerHand := deck[0:17]
	cc.DeductCards(playerHand)
	assert.Equal(t, 37, countTotalCards(cc), "发牌后记牌器37张")

	// 假设玩家2出牌，使用的是 deck[17] 和 deck[18]
	// 这些牌肯定在记牌器中（因为 playerHand 只拿了 0-16）
	otherPlayerPlayed := deck[17:19]
	cc.DeductCards(otherPlayerPlayed)
	assert.Equal(t, 35, countTotalCards(cc), "其他玩家出2张后，记牌器35张")
}

package card

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRankFromChar uses a table-driven test to verify the character-to-rank mapping.
func TestRankFromChar(t *testing.T) {
	t.Parallel()

	// 1. Define the structure for our test cases.
	type testCase struct {
		name         string // A description of the test case.
		inputChar    rune   // The character to pass to the function.
		expectedRank Rank   // The rank we expect to get back.
		expectError  bool   // Whether we expect an error.
	}

	// 2. Create the "table" of test cases.
	testCases := []testCase{
		// --- Valid Uppercase Characters ---
		{name: "valid char '3' for Rank3", inputChar: '3', expectedRank: Rank3, expectError: false},
		{name: "valid char '4' for Rank4", inputChar: '4', expectedRank: Rank4, expectError: false},
		{name: "valid char '5' for Rank5", inputChar: '5', expectedRank: Rank5, expectError: false},
		{name: "valid char '6' for Rank6", inputChar: '6', expectedRank: Rank6, expectError: false},
		{name: "valid char '7' for Rank7", inputChar: '7', expectedRank: Rank7, expectError: false},
		{name: "valid char '8' for Rank8", inputChar: '8', expectedRank: Rank8, expectError: false},
		{name: "valid char '9' for Rank9", inputChar: '9', expectedRank: Rank9, expectError: false},
		{name: "valid char 'T' for Rank10", inputChar: 'T', expectedRank: Rank10, expectError: false},
		{name: "valid char 'J' for RankJ", inputChar: 'J', expectedRank: RankJ, expectError: false},
		{name: "valid char 'Q' for RankQ", inputChar: 'Q', expectedRank: RankQ, expectError: false},
		{name: "valid char 'K' for RankK", inputChar: 'K', expectedRank: RankK, expectError: false},
		{name: "valid char 'A' for RankA", inputChar: 'A', expectedRank: RankA, expectError: false},
		{name: "valid char '2' for Rank2", inputChar: '2', expectedRank: Rank2, expectError: false},
		{name: "valid char 'B' for Black Joker", inputChar: 'B', expectedRank: RankBlackJoker, expectError: false},
		{name: "valid char 'R' for Red Joker", inputChar: 'R', expectedRank: RankRedJoker, expectError: false},

		// --- Invalid Characters (Edge Cases) ---
		{name: "invalid char '1'", inputChar: '1', expectedRank: -1, expectError: true},
		{name: "invalid char '0'", inputChar: '0', expectedRank: -1, expectError: true},
		{name: "invalid char 'X'", inputChar: 'X', expectedRank: -1, expectError: true},
		{name: "invalid lowercase char 'a'", inputChar: 'a', expectedRank: -1, expectError: true},
		{name: "invalid symbol char '$'", inputChar: '$', expectedRank: -1, expectError: true},
	}

	// 3. Loop through the test cases and run the assertions.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Act: Call the function we are testing.
			actualRank, err := RankFromChar(tc.inputChar)

			// Assert: Check the results against our expectations.
			if tc.expectError {
				assert.Error(t, err, "Expected an error for invalid input")
			} else {
				assert.NoError(t, err, "Did not expect an error for valid input")
			}

			// Always check the returned rank, even on error, to ensure it's the default.
			assert.Equal(t, tc.expectedRank, actualRank, "The returned rank should match the expected rank")
		})
	}
}

// TestNewDeck 验证 NewDeck 是否能正确创建一副完整的、无重复的54张牌
func TestNewDeck(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func testRuleCards(ranks ...Rank) []Card {
	cards := make([]Card, len(ranks))
	for i, r := range ranks {
		cards[i] = Card{Rank: r}
	}
	return cards
}

// TestFindCardsInHand uses a table to verify card finding logic.
func TestFindCardsInHand(t *testing.T) {
	t.Parallel()

	// Create a diverse hand for testing multiple scenarios
	fullHand := []Card{
		{Rank: Rank3, Suit: Spade},
		{Rank: Rank4, Suit: Spade},
		{Rank: Rank5, Suit: Spade},
		{Rank: Rank6, Suit: Spade},
		{Rank: Rank7, Suit: Spade},
		{Rank: Rank8, Suit: Heart},
		{Rank: Rank8, Suit: Club},
		{Rank: Rank10, Suit: Club},
		{Rank: RankK, Suit: Diamond},
		{Rank: RankK, Suit: Heart},
		{Rank: RankK, Suit: Spade},
		{Rank: RankA, Suit: Heart},
		{Rank: RankA, Suit: Club},
		{Rank: Rank2, Suit: Heart},
		{Rank: Rank2, Suit: Club},
		{Rank: RankBlackJoker},
		{Rank: RankRedJoker},
	}

	testCases := []struct {
		name          string
		hand          []Card
		input         string
		expectError   bool
		expectedCards []Card // The exact cards we expect to find
	}{
		// --- Success Cases ---
		{
			name:          "find a single card",
			hand:          fullHand,
			input:         "3",
			expectError:   false,
			expectedCards: []Card{{Rank: Rank3, Suit: Spade}},
		},
		{
			name:          "find a pair from a trio",
			hand:          fullHand,
			input:         "KK",
			expectError:   false,
			expectedCards: []Card{{Rank: RankK, Suit: Heart}, {Rank: RankK, Suit: Spade}},
		},
		{
			name:          "find a straight",
			hand:          fullHand,
			input:         "34567",
			expectError:   false,
			expectedCards: testRuleCards(Rank3, Rank4, Rank5, Rank6, Rank7),
		},
		{
			name:          "find a hand with '10'",
			hand:          fullHand,
			input:         "8810",
			expectError:   false,
			expectedCards: []Card{{Rank: Rank8, Suit: Heart}, {Rank: Rank8, Suit: Club}, {Rank: Rank10, Suit: Club}},
		},
		{
			name:          "find the Rocket with JOKER keyword",
			hand:          fullHand,
			input:         "JOKER",
			expectError:   false,
			expectedCards: testRuleCards(RankBlackJoker, RankRedJoker),
		},
		{
			name:          "find the Rocket with RB keyword",
			hand:          fullHand,
			input:         "RB",
			expectError:   false,
			expectedCards: testRuleCards(RankRedJoker, RankBlackJoker),
		},

		// --- Failure Cases ---
		{
			name:          "fail when cards are not in hand",
			hand:          fullHand,
			input:         "QQ",
			expectError:   true,
			expectedCards: nil,
		},
		{
			name:          "fail when not enough cards are in hand",
			hand:          fullHand, // Has three Kings
			input:         "KKKK",
			expectError:   true,
			expectedCards: nil,
		},
		{
			name:          "fail to find Rocket when one is missing",
			hand:          testRuleCards(RankRedJoker, Rank3),
			input:         "JOKER",
			expectError:   true,
			expectedCards: nil,
		},
		{
			name:          "fail with invalid character in input",
			hand:          fullHand,
			input:         "345X",
			expectError:   true,
			expectedCards: nil,
		},

		// --- Edge Cases ---
		{
			name:          "find from an empty hand",
			hand:          []Card{},
			input:         "3",
			expectError:   true,
			expectedCards: nil,
		},
		{
			name:          "find with an empty input string",
			hand:          fullHand,
			input:         "",
			expectError:   false,
			expectedCards: []Card{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actualCards, err := FindCardsInHand(tc.hand, tc.input)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Use ElementsMatch because the order of returned cards is not guaranteed,
				// but the elements themselves should be the same.
				assert.ElementsMatch(t, tc.expectedCards, actualCards)
			}
		})
	}
}

// TestRemoveCards uses a table to verify card removal logic.
func TestRemoveCards(t *testing.T) {
	t.Parallel()

	// Create specific card instances to test exact object removal
	threeOfSpades := Card{Rank: Rank3, Suit: Spade}
	threeOfHearts := Card{Rank: Rank3, Suit: Heart}
	fourOfClubs := Card{Rank: Rank4, Suit: Club}
	kingOfSpades := Card{Rank: RankK, Suit: Spade}

	testCases := []struct {
		name          string
		initialHand   []Card
		cardsToRemove []Card
		expectedHand  []Card
	}{
		{
			name:          "remove a single card",
			initialHand:   []Card{threeOfSpades, fourOfClubs, kingOfSpades},
			cardsToRemove: []Card{fourOfClubs},
			expectedHand:  []Card{threeOfSpades, kingOfSpades},
		},
		{
			name:          "remove multiple cards",
			initialHand:   []Card{threeOfSpades, fourOfClubs, kingOfSpades},
			cardsToRemove: []Card{threeOfSpades, kingOfSpades},
			expectedHand:  []Card{fourOfClubs},
		},
		{
			name:          "remove one of two identical ranks",
			initialHand:   []Card{threeOfSpades, threeOfHearts, fourOfClubs},
			cardsToRemove: []Card{threeOfSpades},
			expectedHand:  []Card{threeOfHearts, fourOfClubs},
		},
		{
			name:          "attempt to remove a card not in hand",
			initialHand:   []Card{threeOfSpades, fourOfClubs},
			cardsToRemove: []Card{kingOfSpades},
			expectedHand:  []Card{threeOfSpades, fourOfClubs},
		},
		{
			name:          "remove all cards",
			initialHand:   []Card{threeOfSpades, fourOfClubs},
			cardsToRemove: []Card{threeOfSpades, fourOfClubs},
			expectedHand:  []Card{},
		},
		{
			name:          "remove from an empty hand",
			initialHand:   []Card{},
			cardsToRemove: []Card{threeOfSpades},
			expectedHand:  []Card{},
		},
		{
			name:          "remove an empty slice of cards",
			initialHand:   []Card{threeOfSpades, fourOfClubs},
			cardsToRemove: []Card{},
			expectedHand:  []Card{threeOfSpades, fourOfClubs},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actualHand := RemoveCards(tc.initialHand, tc.cardsToRemove)
			// ElementsMatch is perfect here as well, in case the order changes.
			assert.ElementsMatch(t, tc.expectedHand, actualHand)
		})
	}
}

// FuzzRankFromChar 对 RankFromChar 函数进行模糊测试
func FuzzRankFromChar(f *testing.F) {
	// 添加种子语料库。这些是有效的、我们期望函数能够正确处理的输入。
	// 模糊测试引擎会使用这些输入作为起点来生成新的、随机的输入。
	f.Add("3")
	f.Add("4")
	f.Add("5")
	f.Add("6")
	f.Add("7")
	f.Add("8")
	f.Add("9")
	f.Add("T")
	f.Add("J")
	f.Add("Q")
	f.Add("K")
	f.Add("A")
	f.Add("2")
	f.Add("B")
	f.Add("R")

	// 模糊测试的目标函数
	f.Fuzz(func(t *testing.T, input string) {
		// 期望至少一个字符
		if len(input) == 0 {
			return
		}

		// 只取第一个 rune（Unicode 码点）进行测试
		r, _ := utf8.DecodeRuneInString(input)
		rank, err := RankFromChar(r)

		// 检查函数是否会因为某些输入而崩溃
		if err != nil { // 如果有错误，选择性地检查 Rank 是否为预期的错误值
			if rank != -1 {
				t.Errorf("对于输入 '%s'，在返回错误时，Rank 应该是 -1，但得到的是 %d", input, rank)
			}
		} else { // 如果没有错误，检查 Rank 是否在有效范围内
			if rank < Rank3 || rank > RankRedJoker {
				t.Errorf("对于输入 '%s'，返回了无效的 Rank: %d", input, rank)
			}
		}
	})
}

// FuzzFindCardsInHand 对 FindCardsInHand 函数进行模糊测试
func FuzzFindCardsInHand(f *testing.F) {
	// 添加一些种子语料库
	// 格式：用一个特殊的分隔符（比如 |）来分隔手牌和输入字符串
	f.Add("34567|345") // 正常情况
	f.Add("AKQJ10|BR") // 王炸检查
	f.Add("B|BR")      // 手牌不全，无法出王炸
	f.Add("333|3333")  // 手牌不够
	f.Add("|345")      // 空手牌
	f.Add("345|")      // 空输入
	f.Add("345|345X")  // 无效输入

	f.Fuzz(func(t *testing.T, data string) {
		// 将模糊测试生成的字符串拆分为手牌部分和出牌部分
		parts := strings.SplitN(data, "|", 2)
		if len(parts) != 2 {
			return // 如果格式不对，就跳过
		}
		handStr, inputStr := parts[0], parts[1]

		var hand []Card
		fullDeck := NewDeck()

		deckMap := make(map[Rank]Card)
		for _, card := range fullDeck {
			if card.Rank == RankBlackJoker || card.Rank == RankRedJoker {
				deckMap[card.Rank] = card
			} else if _, ok := deckMap[card.Rank]; !ok {
				// 对于普通牌，只存一张代表即可
				deckMap[card.Rank] = card
			}
		}

		for _, char := range handStr {
			rank, err := RankFromChar(char)
			if err == nil {
				if card, ok := deckMap[rank]; ok {
					hand = append(hand, card)
				}
			}
		}

		// 只关心它是否会 panic，所以不需要对返回值做过多断言
		_, _ = FindCardsInHand(hand, inputStr)
	})
}

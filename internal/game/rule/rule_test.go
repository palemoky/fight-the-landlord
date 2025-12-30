package rule

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
)

// testRuleCards is a helper to quickly create card slices for testing.
func testRuleCards(ranks ...card.Rank) []card.Card {
	cards := make([]card.Card, len(ranks))
	for i, r := range ranks {
		// Suit and color are irrelevant for rule parsing.
		cards[i] = card.Card{Rank: r}
	}
	return cards
}

func TestIsContinuous(t *testing.T) {
	testCases := []struct {
		name     string
		ranks    []card.Rank
		expected bool
	}{
		{"valid 5-card straight", []card.Rank{card.Rank3, card.Rank4, card.Rank5, card.Rank6, card.Rank7}, true},
		{"valid 3-rank sequence", []card.Rank{card.RankJ, card.RankQ, card.RankK}, true},
		{"not continuous", []card.Rank{card.Rank3, card.Rank4, card.Rank6}, false},
		{"contains invalid Rank2", []card.Rank{card.RankA, card.Rank2}, false},
		{"contains joker rank", []card.Rank{card.RankK, card.RankA, card.RankBlackJoker}, false},
		{"empty slice", []card.Rank{}, false},
		{"single rank", []card.Rank{card.Rank5}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, isContinuous(tc.ranks))
		})
	}
}

func TestAnalyzeCards(t *testing.T) {
	// A complex hand to test all categories
	hand := testRuleCards(
		card.Rank3,             // one
		card.Rank4, card.Rank4, // pair
		card.Rank5, card.Rank5, card.Rank5, // trio
		card.Rank6, card.Rank6, card.Rank6, card.Rank6, // four
		card.RankJ, // one
	)

	analysis := analyzeCards(hand)

	assert.Equal(t, []card.Rank{card.Rank6}, analysis.fours, "Should correctly identify fours and sort them")
	assert.Equal(t, []card.Rank{card.Rank5}, analysis.trios, "Should correctly identify trios and sort them")
	assert.Equal(t, []card.Rank{card.Rank4}, analysis.pairs, "Should correctly identify pairs and sort them")
	assert.Equal(t, []card.Rank{card.Rank3, card.RankJ}, analysis.ones, "Should correctly identify ones and sort them")
	assert.Equal(t, 2, analysis.counts[card.Rank4], "Counts map should be accurate")
}

func TestParseHand(t *testing.T) {
	testCases := []struct {
		name         string
		cards        []card.Card
		expectError  bool
		expectedType HandType
		expectedRank card.Rank
		expectedLen  int
	}{
		// Simple Types
		{"Single 5", testRuleCards(card.Rank5), false, Single, card.Rank5, 0},
		{"Pair 7", testRuleCards(card.Rank7, card.Rank7), false, Pair, card.Rank7, 0},
		{"Trio 9", testRuleCards(card.Rank9, card.Rank9, card.Rank9), false, Trio, card.Rank9, 0},

		// Bomb & Rocket
		{"Rocket", testRuleCards(card.RankBlackJoker, card.RankRedJoker), false, Rocket, card.RankRedJoker, 0},
		{"Bomb of 8s", testRuleCards(card.Rank8, card.Rank8, card.Rank8, card.Rank8), false, Bomb, card.Rank8, 0},

		// Trio with Kickers
		{"Trio with Single", testRuleCards(card.Rank3, card.Rank3, card.Rank3, card.Rank4), false, TrioWithSingle, card.Rank3, 0},
		{"Trio with Pair", testRuleCards(card.RankA, card.RankA, card.RankA, card.RankK, card.RankK), false, TrioWithPair, card.RankA, 0},

		// Straights
		{"5-card Straight", testRuleCards(card.Rank3, card.Rank4, card.Rank5, card.Rank6, card.Rank7), false, Straight, card.Rank3, 5},
		{"3-pair Straight", testRuleCards(card.Rank8, card.Rank8, card.Rank9, card.Rank9, card.Rank10, card.Rank10), false, PairStraight, card.Rank8, 3},

		// Planes
		{"Plane no wings", testRuleCards(card.Rank3, card.Rank3, card.Rank3, card.Rank4, card.Rank4, card.Rank4), false, Plane, card.Rank3, 2},
		{"Plane with singles", testRuleCards(card.Rank5, card.Rank5, card.Rank5, card.Rank6, card.Rank6, card.Rank6, card.Rank7, card.Rank8), false, PlaneWithSingles, card.Rank5, 2},
		{"Plane with pairs", testRuleCards(card.Rank9, card.Rank9, card.Rank9, card.Rank10, card.Rank10, card.Rank10, card.RankJ, card.RankJ, card.RankQ, card.RankQ), false, PlaneWithPairs, card.Rank9, 2},

		// Four with Kickers
		{"Four with two singles", testRuleCards(card.Rank4, card.Rank4, card.Rank4, card.Rank4, card.Rank5, card.Rank6), false, FourWithTwo, card.Rank4, 0},
		{"Four with one pair", testRuleCards(card.Rank4, card.Rank4, card.Rank4, card.Rank4, card.Rank5, card.Rank5), false, FourWithTwo, card.Rank4, 0},
		{"Four with two pairs", testRuleCards(card.RankJ, card.RankJ, card.RankJ, card.RankJ, card.RankQ, card.RankQ, card.RankK, card.RankK), false, FourWithTwoPairs, card.RankJ, 0},

		// Invalid hands
		{"Invalid Trio with Pair", testRuleCards(card.Rank3, card.Rank3, card.Rank3, card.Rank4, card.Rank5), true, Invalid, 0, 0},
		{"Invalid Straight with 2", testRuleCards(card.RankK, card.RankA, card.Rank2, card.Rank3, card.Rank4), true, Invalid, 0, 0},
		{"Empty hand", []card.Card{}, true, Invalid, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parsedHand, err := ParseHand(tc.cards)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedType, parsedHand.Type, "Hand type should match")
				assert.Equal(t, tc.expectedRank, parsedHand.KeyRank, "Key rank should match")
				if tc.expectedLen > 0 {
					assert.Equal(t, tc.expectedLen, parsedHand.Length, "Length should match")
				}
			}
		})
	}
}

func TestCanBeat(t *testing.T) {
	// Helper to quickly create a parsed hand for testing
	ph := func(ht HandType, kr card.Rank, l int) ParsedHand {
		return ParsedHand{Type: ht, KeyRank: kr, Length: l}
	}

	testCases := []struct {
		name     string
		newHand  ParsedHand
		lastHand ParsedHand
		expected bool
	}{
		// Rocket Rules
		{"Rocket beats Bomb", ph(Rocket, card.RankRedJoker, 0), ph(Bomb, card.RankA, 0), true},
		{"Rocket beats anything", ph(Rocket, card.RankRedJoker, 0), ph(Pair, card.Rank2, 0), true},
		{"Anything cannot beat Rocket", ph(Bomb, card.Rank2, 0), ph(Rocket, card.RankRedJoker, 0), false},

		// Bomb Rules
		{"Bomb beats non-bomb", ph(Bomb, card.Rank3, 0), ph(Pair, card.Rank2, 0), true},
		{"Bigger bomb beats smaller bomb", ph(Bomb, card.Rank4, 0), ph(Bomb, card.Rank3, 0), true},
		{"Smaller bomb cannot beat bigger bomb", ph(Bomb, card.Rank5, 0), ph(Bomb, card.Rank6, 0), false},

		// Same Type Rules
		{"Higher pair beats lower pair", ph(Pair, card.RankA, 0), ph(Pair, card.RankK, 0), true},
		{"Lower single cannot beat higher single", ph(Single, card.Rank3, 0), ph(Single, card.Rank4, 0), false},
		{"Higher straight beats lower straight", ph(Straight, card.Rank4, 5), ph(Straight, card.Rank3, 5), true},

		// Mismatched Rules
		{"Mismatched types cannot beat", ph(Pair, card.Rank3, 0), ph(Single, card.RankA, 0), false},
		{"Mismatched length straight cannot beat", ph(Straight, card.Rank3, 6), ph(Straight, card.Rank4, 5), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, CanBeat(tc.newHand, tc.lastHand))
		})
	}
}

// TestCanBeatWithHand verifies the logic for checking if a player has any valid move.
func TestCanBeatWithHand(t *testing.T) {
	// Helper to quickly create a parsed hand for the opponent
	ph := func(ht HandType, kr card.Rank, l int) ParsedHand {
		return ParsedHand{Type: ht, KeyRank: kr, Length: l}
	}

	testCases := []struct {
		name         string
		playerHand   []card.Card
		opponentHand ParsedHand
		expected     bool
	}{
		// --- Trivial Case ---
		{
			name:         "can always play on an empty board",
			playerHand:   testRuleCards(card.Rank3),
			opponentHand: ParsedHand{Type: Invalid}, // IsEmpty() will be true
			expected:     true,
		},

		// --- Rocket Case ---
		{
			name:         "Rocket can beat a Bomb",
			playerHand:   testRuleCards(card.RankBlackJoker, card.RankRedJoker),
			opponentHand: ph(Bomb, card.RankA, 0),
			expected:     true,
		},
		{
			name:         "Rocket can beat a normal hand (Plane)",
			playerHand:   testRuleCards(card.RankBlackJoker, card.RankRedJoker, card.Rank3),
			opponentHand: ph(Plane, card.Rank10, 2),
			expected:     true,
		},

		// --- Bomb Case ---
		{
			name:         "Bomb cannot beat a Rocket",
			playerHand:   testRuleCards(card.Rank2, card.Rank2, card.Rank2, card.Rank2),
			opponentHand: ph(Rocket, card.RankRedJoker, 0),
			expected:     false,
		},
		{
			name:         "Bigger bomb beats smaller bomb",
			playerHand:   testRuleCards(card.Rank5, card.Rank5, card.Rank5, card.Rank5),
			opponentHand: ph(Bomb, card.Rank4, 0),
			expected:     true,
		},
		{
			name:         "Smaller bomb cannot beat bigger bomb",
			playerHand:   testRuleCards(card.Rank3, card.Rank3, card.Rank3, card.Rank3),
			opponentHand: ph(Bomb, card.Rank4, 0),
			expected:     false,
		},
		{
			name:         "Bomb can beat FourWithTwo",
			playerHand:   testRuleCards(card.Rank7, card.Rank7, card.Rank7, card.Rank7),
			opponentHand: ph(FourWithTwo, card.Rank6, 0),
			expected:     true,
		},
		{
			name:         "Bomb can beat TrioWithPair",
			playerHand:   testRuleCards(card.Rank8, card.Rank8, card.Rank8, card.Rank8),
			opponentHand: ph(TrioWithPair, card.RankA, 0),
			expected:     true,
		},
		{
			name:         "Bomb can beat PairStraight",
			playerHand:   testRuleCards(card.Rank9, card.Rank9, card.Rank9, card.Rank9),
			opponentHand: ph(PairStraight, card.RankJ, 3),
			expected:     true,
		},
		{
			name:         "Bomb can beat Single",
			playerHand:   testRuleCards(card.Rank3, card.Rank3, card.Rank3, card.Rank3),
			opponentHand: ph(Single, card.Rank2, 0),
			expected:     true,
		},

		// --- Trivial Case ---
		// --- 单牌 & 对子 ---
		{
			name:         "can beat single with a higher single",
			playerHand:   testRuleCards(card.RankA),
			opponentHand: ph(Single, card.RankQ, 0),
			expected:     true,
		},
		{
			name:         "cannot beat single when no higher single exists",
			playerHand:   testRuleCards(card.Rank3),
			opponentHand: ph(Single, card.RankQ, 0),
			expected:     false,
		},
		{
			name:         "can beat pair with a higher pair",
			playerHand:   testRuleCards(card.RankJ, card.RankJ),
			opponentHand: ph(Pair, card.Rank10, 0),
			expected:     true,
		},
		// --- 三张 & 三带X ---
		{
			name:         "can beat trio with a higher trio",
			playerHand:   testRuleCards(card.Rank5, card.Rank5, card.Rank5),
			opponentHand: ph(Trio, card.Rank4, 0),
			expected:     true,
		},
		{
			name:         "can beat trio with single with a higher trio and a kicker",
			playerHand:   testRuleCards(card.Rank4, card.Rank4, card.Rank4, card.Rank5),
			opponentHand: ph(TrioWithSingle, card.Rank3, 0),
			expected:     true,
		},
		{
			name:         "cannot beat trio with single without a kicker",
			playerHand:   testRuleCards(card.RankK, card.RankK, card.RankK), // has higher trio, but no other card
			opponentHand: ph(TrioWithSingle, card.RankQ, 0),
			expected:     false,
		},
		{
			name:         "can beat trio with pair with a higher trio and a pair kicker",
			playerHand:   testRuleCards(card.Rank7, card.Rank7, card.Rank7, card.Rank8, card.Rank8),
			opponentHand: ph(TrioWithPair, card.Rank6, 0),
			expected:     true,
		},
		{
			name:         "cannot beat trio with pair without a pair kicker",
			playerHand:   testRuleCards(card.RankJ, card.RankJ, card.RankJ, card.Rank3, card.Rank4), // has higher trio, but only single kickers
			opponentHand: ph(TrioWithPair, card.Rank10, 0),
			expected:     false,
		},
		// --- 顺子 & 连对 ---
		{
			name:         "can beat straight with a higher straight",
			playerHand:   testRuleCards(card.Rank4, card.Rank5, card.Rank6, card.Rank7, card.Rank8),
			opponentHand: ph(Straight, card.Rank3, 5),
			expected:     true,
		},
		{
			name:         "cannot beat straight of same length when no higher one exists",
			playerHand:   testRuleCards(card.Rank3, card.Rank4, card.Rank5, card.Rank6, card.Rank7),
			opponentHand: ph(Straight, card.Rank4, 5),
			expected:     false,
		},
		{
			name:         "can beat pair straight with higher pair straight",
			playerHand:   testRuleCards(card.Rank5, card.Rank5, card.Rank6, card.Rank6, card.Rank7, card.Rank7),
			opponentHand: ph(PairStraight, card.Rank4, 3),
			expected:     true,
		},
		// --- 飞机 (Plane) ---
		{
			name:         "can beat plane with pairs with a higher plane and kickers",
			playerHand:   testRuleCards(card.Rank5, card.Rank5, card.Rank5, card.Rank6, card.Rank6, card.Rank6, card.RankA, card.RankA, card.RankK, card.RankK),
			opponentHand: ph(PlaneWithPairs, card.Rank3, 2),
			expected:     true,
		},
		{
			name:         "cannot beat plane when plane is higher but not enough kickers",
			playerHand:   testRuleCards(card.Rank8, card.Rank8, card.Rank8, card.Rank9, card.Rank9, card.Rank9, card.Rank3, card.Rank4), // higher plane, but only single kickers
			opponentHand: ph(PlaneWithPairs, card.Rank6, 2),
			expected:     false,
		},
		{
			name:         "can beat Plane (no wings) with a higher Plane",
			playerHand:   testRuleCards(card.Rank7, card.Rank7, card.Rank7, card.Rank8, card.Rank8, card.Rank8),
			opponentHand: ph(Plane, card.Rank5, 2), // 对手是 555666
			expected:     true,
		},
		{
			name:         "cannot beat Plane (no wings) when no higher one exists",
			playerHand:   testRuleCards(card.Rank3, card.Rank3, card.Rank3, card.Rank4, card.Rank4, card.Rank4, card.RankA), // 有飞机，但不够大
			opponentHand: ph(Plane, card.Rank5, 2),
			expected:     false,
		},
		{
			name:         "can beat PlaneWithSingles with a higher Plane and single kickers",
			playerHand:   testRuleCards(card.Rank4, card.Rank4, card.Rank4, card.Rank5, card.Rank5, card.Rank5, card.RankA, card.RankK), // 444555 + A K
			opponentHand: ph(PlaneWithSingles, card.Rank3, 2),                                                                           // 对手是 333444 + 5 6
			expected:     true,
		},
		{
			name:         "cannot beat PlaneWithSingles when plane is higher but not enough single kickers",
			playerHand:   testRuleCards(card.Rank9, card.Rank9, card.Rank9, card.Rank10, card.Rank10, card.Rank10, card.RankJ), // 999TTT + J (只有1个kicker)
			opponentHand: ph(PlaneWithSingles, card.Rank6, 2),                                                                  // 需要2个kicker
			expected:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Sort the hand as it would be in the game, which might affect some logic.
			sort.Slice(tc.playerHand, func(i, j int) bool { return tc.playerHand[i].Rank < tc.playerHand[j].Rank })

			actual := CanBeatWithHand(tc.playerHand, tc.opponentHand)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

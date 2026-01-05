package card

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInputRanks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[Rank]int
		hasError bool
	}{
		{
			name:  "Single card",
			input: "3",
			expected: map[Rank]int{
				Rank3: 1,
			},
			hasError: false,
		},
		{
			name:  "Pair",
			input: "33",
			expected: map[Rank]int{
				Rank3: 2,
			},
			hasError: false,
		},
		{
			name:  "Multiple ranks",
			input: "345",
			expected: map[Rank]int{
				Rank3: 1,
				Rank4: 1,
				Rank5: 1,
			},
			hasError: false,
		},
		{
			name:  "With 10",
			input: "10JQ",
			expected: map[Rank]int{
				Rank10: 1,
				RankJ:  1,
				RankQ:  1,
			},
			hasError: false,
		},
		{
			name:     "Invalid character",
			input:    "3X5",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parseInputRanks(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCountHandRanks(t *testing.T) {
	t.Parallel()

	hand := []Card{
		{Suit: Spade, Rank: Rank3, Color: Black},
		{Suit: Heart, Rank: Rank3, Color: Red},
		{Suit: Diamond, Rank: Rank3, Color: Red},
		{Suit: Club, Rank: Rank4, Color: Black},
		{Suit: Spade, Rank: Rank4, Color: Black},
	}

	counts := countHandRanks(hand)

	assert.Equal(t, 3, counts[Rank3])
	assert.Equal(t, 2, counts[Rank4])
	assert.Equal(t, 0, counts[Rank5])
}

func TestExtractCards(t *testing.T) {
	t.Parallel()

	hand := []Card{
		{Suit: Spade, Rank: Rank3, Color: Black},
		{Suit: Heart, Rank: Rank3, Color: Red},
		{Suit: Diamond, Rank: Rank4, Color: Red},
		{Suit: Club, Rank: Rank4, Color: Black},
	}

	inputRanks := map[Rank]int{
		Rank3: 2,
		Rank4: 1,
	}

	result := extractCards(hand, inputRanks)

	assert.Len(t, result, 3)

	// Count extracted ranks
	extracted := make(map[Rank]int)
	for _, c := range result {
		extracted[c.Rank]++
	}
	assert.Equal(t, 2, extracted[Rank3])
	assert.Equal(t, 1, extracted[Rank4])
}

func TestFindRocketInHand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		hand  []Card
		found bool
	}{
		{
			name: "Has rocket",
			hand: []Card{
				{Suit: Joker, Rank: RankBlackJoker, Color: Black},
				{Suit: Joker, Rank: RankRedJoker, Color: Red},
				{Suit: Spade, Rank: Rank3, Color: Black},
			},
			found: true,
		},
		{
			name: "Only black joker",
			hand: []Card{
				{Suit: Joker, Rank: RankBlackJoker, Color: Black},
				{Suit: Spade, Rank: Rank3, Color: Black},
			},
			found: false,
		},
		{
			name: "Only red joker",
			hand: []Card{
				{Suit: Joker, Rank: RankRedJoker, Color: Red},
				{Suit: Spade, Rank: Rank3, Color: Black},
			},
			found: false,
		},
		{
			name: "No jokers",
			hand: []Card{
				{Suit: Spade, Rank: Rank3, Color: Black},
				{Suit: Heart, Rank: Rank4, Color: Red},
			},
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cards, found := findRocketInHand(tt.hand)
			assert.Equal(t, tt.found, found)
			if found {
				assert.Len(t, cards, 2)
			} else {
				assert.Nil(t, cards)
			}
		})
	}
}

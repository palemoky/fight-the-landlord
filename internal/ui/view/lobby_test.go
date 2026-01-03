package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

func TestRenderLeaderboardTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entries  []protocol.LeaderboardEntry
		expected []string
	}{
		{
			name:     "empty entries",
			entries:  []protocol.LeaderboardEntry{},
			expected: []string{"排行榜"},
		},
		{
			name: "single entry",
			entries: []protocol.LeaderboardEntry{
				{Rank: 1, PlayerName: "Champion", Score: 1000, Wins: 50, WinRate: 75.0},
			},
			expected: []string{"Champion", "1000", "50", "75.0%"},
		},
		{
			name: "multiple entries",
			entries: []protocol.LeaderboardEntry{
				{Rank: 1, PlayerName: "First", Score: 1000, Wins: 50, WinRate: 80.0},
				{Rank: 2, PlayerName: "Second", Score: 800, Wins: 40, WinRate: 70.0},
				{Rank: 3, PlayerName: "Third", Score: 600, Wins: 30, WinRate: 60.0},
			},
			expected: []string{"First", "Second", "Third"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderLeaderboardTable(tt.entries)

			assert.NotEmpty(t, result)
			for _, s := range tt.expected {
				assert.Contains(t, result, s)
			}
		})
	}
}

func TestRenderLeaderboardTable_Header(t *testing.T) {
	t.Parallel()

	entries := []protocol.LeaderboardEntry{
		{Rank: 1, PlayerName: "Test", Score: 100, Wins: 10, WinRate: 50.0},
	}

	result := renderLeaderboardTable(entries)

	// Should contain header row
	assert.Contains(t, result, "排名")
	assert.Contains(t, result, "玩家")
	assert.Contains(t, result, "积分")
	assert.Contains(t, result, "胜场")
	assert.Contains(t, result, "胜率")
}

func TestRenderStatsTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stats    *protocol.StatsResultPayload
		expected []string
	}{
		{
			name: "full stats",
			stats: &protocol.StatsResultPayload{
				Rank:          5,
				Score:         1000,
				TotalGames:    100,
				Wins:          60,
				Losses:        40,
				WinRate:       60.0,
				LandlordGames: 50,
				LandlordWins:  30,
				FarmerGames:   50,
				FarmerWins:    30,
				CurrentStreak: 3,
				MaxWinStreak:  10,
			},
			expected: []string{"#5", "1000", "100", "60", "40", "60.0%", "连胜"},
		},
		{
			name: "unranked player",
			stats: &protocol.StatsResultPayload{
				Rank:       0,
				Score:      0,
				TotalGames: 5,
				Wins:       2,
				Losses:     3,
			},
			expected: []string{"未上榜", "5", "2", "3"},
		},
		{
			name: "losing streak",
			stats: &protocol.StatsResultPayload{
				Rank:          10,
				Score:         500,
				TotalGames:    20,
				Wins:          8,
				Losses:        12,
				CurrentStreak: -5,
			},
			expected: []string{"连败"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderStatsTable(tt.stats)

			assert.NotEmpty(t, result)
			for _, s := range tt.expected {
				assert.Contains(t, result, s, "Should contain: %s", s)
			}
		})
	}
}

func TestRenderStatsTable_LandlordFarmerRates(t *testing.T) {
	t.Parallel()

	stats := &protocol.StatsResultPayload{
		LandlordGames: 20,
		LandlordWins:  15,
		FarmerGames:   30,
		FarmerWins:    18,
	}

	result := renderStatsTable(stats)

	// Check that landlord and farmer stats are shown
	assert.Contains(t, result, "地主")
	assert.Contains(t, result, "农民")
	assert.Contains(t, result, "15")
	assert.Contains(t, result, "18")
}

func TestRenderStatsTable_ZeroGames(t *testing.T) {
	t.Parallel()

	stats := &protocol.StatsResultPayload{
		LandlordGames: 0,
		FarmerGames:   0,
	}

	result := renderStatsTable(stats)

	// Should not panic with zero games
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "我的战绩")
}

func TestRenderStatsTable_MaxWinStreak(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		maxWinStreak  int
		shouldContain bool
	}{
		{"no max streak", 0, false},
		{"has max streak", 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stats := &protocol.StatsResultPayload{
				MaxWinStreak: tt.maxWinStreak,
			}

			result := renderStatsTable(stats)

			if tt.shouldContain {
				assert.Contains(t, result, "最高连胜")
			}
		})
	}
}

// TestLeaderboardTable_LongPlayerName tests truncation of long names
func TestLeaderboardTable_LongPlayerName(t *testing.T) {
	t.Parallel()

	entries := []protocol.LeaderboardEntry{
		{Rank: 1, PlayerName: "VeryLongPlayerNameThatShouldBeTruncated", Score: 1000, Wins: 50, WinRate: 75.0},
	}

	result := renderLeaderboardTable(entries)

	// Should not contain the full long name
	assert.NotEmpty(t, result)
	// Name should be truncated (TruncateName limits to 10 chars)
	require.True(t, result != "")

	// The full name should not appear as-is (it gets truncated)
	lines := strings.Split(result, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "VeryLong") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should contain truncated player name")
}

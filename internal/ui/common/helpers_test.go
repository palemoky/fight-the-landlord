package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short name within limit", "Alice", 10, "Alice"},
		{"exact length", "HelloWorld", 10, "HelloWorld"},
		{"long name truncated", "VeryLongPlayerName", 10, "VeryLongP…"},
		{"chinese name truncated", "可爱的龙猫", 4, "可爱的…"},
		{"empty name", "", 10, ""},
		{"emoji handling", "🎮玩家名字很长", 5, "🎮玩家名…"},
		{"single char limit", "Hello", 1, "…"},
		{"unicode mixed exact", "Hello世界", 7, "Hello世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := TruncateName(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

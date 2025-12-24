package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short name no truncation", "Alice", 10, "Alice"},
		{"exact length no truncation", "Alice", 5, "Alice"},
		{"long name truncated", "AliceInWonderland", 10, "AliceInWo…"},
		{"unicode characters", "玩家一二三四五", 5, "玩家一二…"},
		{"mixed unicode and ascii truncated", "Player玩家名", 8, "Player玩…"},
		{"empty string", "", 5, ""},
		{"single character", "A", 5, "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, truncateName(tt.input, tt.maxLen))
		})
	}
}

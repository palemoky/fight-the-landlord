package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderQuickMsgMenu(t *testing.T) {
	t.Parallel()

	result := RenderQuickMsgMenu()

	// Should contain the header
	assert.Contains(t, result, "快捷消息")
	assert.Contains(t, result, "数字键选择")

	// Should contain numbered messages
	assert.Contains(t, result, "1.")
	assert.Contains(t, result, QuickMessages[0])

	// Should not be empty
	assert.NotEmpty(t, result)
}

func TestQuickMessages(t *testing.T) {
	t.Parallel()

	// QuickMessages should have reasonable count
	assert.Greater(t, len(QuickMessages), 10, "Should have at least 10 quick messages")
	assert.LessOrEqual(t, len(QuickMessages), 30, "Should not have too many messages")

	// Each message should not be empty
	for i, msg := range QuickMessages {
		assert.NotEmpty(t, msg, "Message %d should not be empty", i+1)
	}
}

func TestRenderChatBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		history          []string
		expectedEmpty    bool
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "empty history returns empty",
			history:       []string{},
			expectedEmpty: true,
		},
		{
			name:          "nil history returns empty",
			history:       nil,
			expectedEmpty: true,
		},
		{
			name:    "single message",
			history: []string{"[12:00] Player: Hello"},
			shouldContain: []string{
				"Player",
				"Hello",
			},
		},
		{
			name: "multiple messages under limit",
			history: []string{
				"[12:00] A: msg1",
				"[12:01] B: msg2",
				"[12:02] C: msg3",
			},
			shouldContain: []string{
				"msg1",
				"msg2",
				"msg3",
			},
		},
		{
			name: "more than 5 messages shows only last 5",
			history: []string{
				"[12:00] A: oldest",
				"[12:01] B: old2",
				"[12:02] C: old3",
				"[12:03] D: msg4",
				"[12:04] E: msg5",
				"[12:05] F: msg6",
				"[12:06] G: newest",
			},
			shouldContain: []string{
				"msg4",
				"msg5",
				"msg6",
				"newest",
			},
			shouldNotContain: []string{
				"oldest",
				"old2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := RenderChatBox(tt.history)

			if tt.expectedEmpty {
				assert.Empty(t, result)
				return
			}

			assert.NotEmpty(t, result)

			for _, s := range tt.shouldContain {
				assert.Contains(t, result, s, "Should contain: %s", s)
			}

			for _, s := range tt.shouldNotContain {
				assert.NotContains(t, result, s, "Should not contain: %s", s)
			}
		})
	}
}

func TestRenderChatBox_MessageCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		historyLen    int
		expectedShown int
	}{
		{"1 message", 1, 1},
		{"3 messages", 3, 3},
		{"5 messages", 5, 5},
		{"7 messages", 7, 5},
		{"10 messages", 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			history := make([]string, tt.historyLen)
			for i := range history {
				history[i] = "msg"
			}

			result := RenderChatBox(history)
			// Count occurrences of "msg" lines
			count := strings.Count(result, "msg")
			assert.Equal(t, tt.expectedShown, count, "Should show %d messages", tt.expectedShown)
		})
	}
}

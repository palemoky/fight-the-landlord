package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderGameRules(t *testing.T) {
	t.Parallel()

	result := RenderGameRules()

	tests := []struct {
		name     string
		contains string
	}{
		{"game goal section", "【游戏目标】"},
		{"landlord rule", "地主"},
		{"farmer rule", "农民"},
		{"card type section", "【牌型说明】"},
		{"single card", "单牌"},
		{"pair", "对子"},
		{"trio", "三张"},
		{"trio with single", "三带一"},
		{"trio with pair", "三带二"},
		{"straight", "顺子"},
		{"pair straight", "连对"},
		{"plane", "飞机"},
		{"four with two", "四带二"},
		{"bomb", "炸弹"},
		{"rocket", "王炸"},
		{"bidding section", "【叫地主规则】"},
		{"play rules section", "【出牌规则】"},
		{"shortcut section", "【快捷键】"},
		{"toggle counter key", "C："},
		{"toggle message key", "T："},
		{"help key", "H："},
		{"escape key", "ESC："},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, result, tt.contains, "Should contain: %s", tt.contains)
		})
	}
}

func TestRenderGameRules_NotEmpty(t *testing.T) {
	t.Parallel()

	result := RenderGameRules()

	assert.NotEmpty(t, result)
	// Should have substantial content (more than 100 chars for rules)
	assert.Greater(t, len(result), 100)
}

func TestRulesView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard size", 80, 24},
		{"wide screen", 120, 40},
		{"small screen", 60, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := RulesView(tt.width, tt.height)

			assert.NotEmpty(t, result)
			// Should contain the title
			assert.Contains(t, result, "游戏规则")
			// Should contain actual game rules
			assert.True(t, strings.Contains(result, "地主") || strings.Contains(result, "农民"))
		})
	}
}

func TestRulesView_ContainsTitle(t *testing.T) {
	t.Parallel()

	result := RulesView(80, 24)

	// Title should be present
	assert.Contains(t, result, "📖")
	assert.Contains(t, result, "游戏规则")
}

package ui

import (
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// --- 辅助函数 ---

// parseCardsInput 解析出牌输入
func (m *OnlineModel) parseCardsInput(input string) ([]card.Card, error) {
	return card.FindCardsInHand(m.game.hand, strings.ToUpper(input))
}

// sortHand 排序手牌
func (m *OnlineModel) sortHand() {
	sort.Slice(m.game.hand, func(i, j int) bool {
		return m.game.hand[i].Rank > m.game.hand[j].Rank
	})
}

// resetGameState 重置游戏状态
func (m *OnlineModel) resetGameState() {
	m.game.roomCode = ""
	m.game.players = nil
	m.game.hand = nil
	m.game.landlordCards = nil
	m.game.currentTurn = ""
	m.game.lastPlayedBy = ""
	m.game.lastPlayed = nil
	m.game.isLandlord = false
	m.game.winner = ""
	m.input.Placeholder = "↑↓ 选择 | 回车确认 | 或输入房间号"
}

// restoreGameState 从重连数据恢复游戏状态
func (m *OnlineModel) restoreGameState(gs *protocol.GameStateDTO) {
	// deprecated: use restoreGameState in online_handlers.go
}

// --- 提醒音相关 ---

// shouldPlayBell 判断是否应该播放提示音
func (m *OnlineModel) shouldPlayBell() bool {
	// 已经播放过了
	if m.game.bellPlayed {
		return false
	}

	// 必须是自己的回合
	isMyTurn := (m.phase == PhaseBidding && m.game.bidTurn == m.playerID) ||
		(m.phase == PhasePlaying && m.game.currentTurn == m.playerID)
	if !isMyTurn {
		return false
	}

	// 检查剩余时间是否为 10 秒
	if m.game.timerStartTime.IsZero() {
		return false
	}

	elapsed := time.Since(m.game.timerStartTime)
	remaining := m.game.timerDuration - elapsed

	return remaining <= 10*time.Second && remaining > 9*time.Second
}

// playBell 播放终端提示音
func (m *OnlineModel) playBell() tea.Cmd {
	return tea.Printf("\a") // 发送 ASCII Bell 字符
}

// resetBell 重置提示音状态（新回合开始时调用）
func (m *OnlineModel) resetBell() {
	m.game.bellPlayed = false
}

// --- 工具函数 ---

// truncateName 截断玩家名称
func truncateName(name string, maxLen int) string {
	runes := []rune(name)
	if len(runes) > maxLen {
		return string(runes[:maxLen-1]) + "…"
	}
	return name
}

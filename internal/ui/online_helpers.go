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
	return card.FindCardsInHand(m.hand, strings.ToUpper(input))
}

// sortHand 排序手牌
func (m *OnlineModel) sortHand() {
	sort.Slice(m.hand, func(i, j int) bool {
		return m.hand[i].Rank > m.hand[j].Rank
	})
}

// resetGameState 重置游戏状态
func (m *OnlineModel) resetGameState() {
	m.roomCode = ""
	m.players = nil
	m.hand = nil
	m.landlordCards = nil
	m.currentTurn = ""
	m.lastPlayedBy = ""
	m.lastPlayed = nil
	m.isLandlord = false
	m.winner = ""
	m.input.Placeholder = "1=创建房间, 2=加入房间, 3=快速匹配"
}

// restoreGameState 从重连数据恢复游戏状态
func (m *OnlineModel) restoreGameState(gs *protocol.GameStateDTO) {
	m.players = gs.Players
	m.hand = protocol.InfosToCards(gs.Hand)
	m.sortHand()
	m.landlordCards = protocol.InfosToCards(gs.LandlordCards)
	m.currentTurn = gs.CurrentTurn
	m.lastPlayed = protocol.InfosToCards(gs.LastPlayed)
	m.lastPlayedBy = gs.LastPlayerID
	m.mustPlay = gs.MustPlay
	m.canBeat = gs.CanBeat

	// 找出自己是否是地主
	for _, p := range m.players {
		if p.ID == m.playerID && p.IsLandlord {
			m.isLandlord = true
			break
		}
	}

	// 根据阶段设置 phase
	switch gs.Phase {
	case "bidding":
		m.phase = PhaseBidding
	case "playing":
		m.phase = PhasePlaying
	case "ended":
		m.phase = PhaseGameOver
	default:
		m.phase = PhaseWaiting
	}
}

// --- 提醒音相关 ---

// shouldPlayBell 判断是否应该播放提示音
func (m *OnlineModel) shouldPlayBell() bool {
	// 已经播放过了
	if m.bellPlayed {
		return false
	}

	// 必须是自己的回合
	isMyTurn := (m.phase == PhaseBidding && m.bidTurn == m.playerID) ||
		(m.phase == PhasePlaying && m.currentTurn == m.playerID)
	if !isMyTurn {
		return false
	}

	// 检查剩余时间是否为 10 秒
	if m.timerStartTime.IsZero() {
		return false
	}

	elapsed := time.Since(m.timerStartTime)
	remaining := m.timerDuration - elapsed

	return remaining <= 10*time.Second && remaining > 9*time.Second
}

// playBell 播放终端提示音
func (m *OnlineModel) playBell() tea.Cmd {
	return tea.Printf("\a") // 发送 ASCII Bell 字符
}

// resetBell 重置提示音状态（新回合开始时调用）
func (m *OnlineModel) resetBell() {
	m.bellPlayed = false
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

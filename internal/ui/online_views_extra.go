package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// connectingView 显示连接中状态
func (m *OnlineModel) connectingView() string {
	var sb string
	if m.error != "" {
		sb = errorStyle.Render(m.error)
	} else {
		sb = "正在连接服务器..."
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb)
}

// matchingView 显示匹配中状态
func (m *OnlineModel) matchingView() string {
	elapsed := time.Since(m.matchingStartTime).Seconds()
	msg := fmt.Sprintf("🔍 正在匹配玩家...\n\n已等待: %.0f 秒\n\n按 ESC 取消", elapsed)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}

// Package view provides UI rendering functions.
package view

import (
	"fmt"
	"strings"
)

// QuickMessages is the list of quick messages for in-game chat.
var QuickMessages = []string{
	"时间不多了哦～",
	"别想了，出吧！",
	"天亮前能打完吗？",
	"你这是读我牌了吗？",
	"这也太倒霉了吧！",
	"有点东西啊！",
	"好家伙！我裂开了...",
	"给我一次机会吧！",
	"手气不错，今晚该买彩票了！",
	"我感觉胜利在向我招手~",
	"这局有点刺激啊！",
	"这牌……是系统针对我吧？",
	"这牌我都不好意思打...",
	"命运对我下手太狠了...",
	"淡定，牌是慢慢变好的。",
	"坐稳了，这把我来！",
	"让你们见识一下技术！",
	"我赌你接不住这一手！",
	"配合得不错，点赞 👍",
	"今晚就到这吧～",
	"能遇到你们真开心！",
	"输了也开心，玩得舒服！",
	"输赢不重要，开心最重要！",
}

// RenderQuickMsgMenu renders the quick message menu.
func RenderQuickMsgMenu() string {
	var sb strings.Builder
	sb.WriteString("💬 快捷消息 (数字键选择)\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	for i, msg := range QuickMessages {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, msg)
	}

	return BoxStyle.Render(sb.String())
}

// RenderChatBox renders the chat box for game view.
func RenderChatBox(history []string) string {
	if len(history) == 0 {
		return ""
	}

	var chatBuilder strings.Builder
	count := len(history)
	start := 0
	if count > 5 {
		start = count - 5
	}
	for i := start; i < count; i++ {
		chatBuilder.WriteString(history[i] + "\n")
	}
	return BoxStyle.Width(40).Render(chatBuilder.String())
}

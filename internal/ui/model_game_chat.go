package ui

import (
	"fmt"
	"strings"
)

var quickMessages = []string{
	"快点啊，我等的花儿都谢了！",
	"你的牌打得也太好了！",
	"交个朋友吧，能告诉我你的联系方式吗？",
	"大家好，很高兴见到各位！",
	"和你合作真是太愉快了！",
	"不要走，决战到天亮！",
	"再见了，我会想念大家的！",
	"这也太倒霉了吧！",
}

func (m *GameModel) renderQuickMsgMenu() string {
	var sb strings.Builder
	sb.WriteString("💬 快捷消息 (数字键选择)\n")
	sb.WriteString(strings.Repeat("─", 30) + "\n")

	for i, msg := range quickMessages {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, msg))
	}
	sb.WriteString(strings.Repeat("─", 30) + "\n")
	sb.WriteString("ESC 关闭")

	return boxStyle.Render(sb.String())
}

func (m *GameModel) renderChatBox() string {
	if len(m.chatHistory) == 0 {
		return ""
	}

	var chatBuilder strings.Builder
	count := len(m.chatHistory)
	start := 0
	if count > 5 {
		start = count - 5
	}
	for i := start; i < count; i++ {
		chatBuilder.WriteString(m.chatHistory[i] + "\n")
	}
	return boxStyle.Width(40).Render(chatBuilder.String())
}

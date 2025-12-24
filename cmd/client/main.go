package main

import (
	"flag"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/ui"
)

func main() {
	serverAddr := flag.String("server", "localhost:1780", "服务器地址")
	flag.Parse()

	serverURL := fmt.Sprintf("ws://%s/ws", *serverAddr)

	model := ui.NewOnlineModel(serverURL)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("启动客户端时出错: %v", err)
	}
}

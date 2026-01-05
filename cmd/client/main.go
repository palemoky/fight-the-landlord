package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/logger"
	"github.com/palemoky/fight-the-landlord/internal/ui"
)

// 默认服务器地址（可通过编译时 -ldflags 注入）
var defaultServer = "localhost:1780"

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Panic recovery to ensure terminal state is restored
	defer func() {
		if r := recover(); r != nil {
			// Log panic to file
			logger.LogPanic(r)

			// Clear screen and restore terminal
			fmt.Print("\033[2J\033[H") // Clear screen
			fmt.Print("\033[?25h")     // Show cursor
			fmt.Fprintf(os.Stderr, "\n[PANIC] 客户端崩溃: %v\n\n", r)
			fmt.Fprintf(os.Stderr, "详细日志已保存到: %s\n", logger.GetLogPath())
			os.Exit(1)
		}
	}()

	serverAddr := flag.String("server", defaultServer, "服务器地址")
	flag.Parse()

	// 支持完整 URL (wss://...) 或仅 host:port
	var serverURL string
	if strings.HasPrefix(*serverAddr, "ws://") || strings.HasPrefix(*serverAddr, "wss://") {
		serverURL = *serverAddr
	} else {
		serverURL = fmt.Sprintf("ws://%s/ws", *serverAddr)
	}
	logger.LogInfo("Connecting to server: %s", serverURL)

	model := ui.NewOnlineModel(serverURL)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logger.LogError("Client error: %v", err)
		log.Printf("启动客户端时出错: %v", err)
	}
}

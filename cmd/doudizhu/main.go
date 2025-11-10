package main

import (
	"github.com/palemoky/fight-the-landlord-go/internal/ui"
	"github.com/palemoky/fight-the-landlord-go/internal/game"
)

func main() {
	// 创建一个UI实例
	terminalUI := ui.NewTerminalUI()

	// 创建游戏实例，并将 UI 注入进去
	doudizhuGame := game.NewGame(terminalUI)

	// 启动游戏
	doudizhuGame.Run()
}

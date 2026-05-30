package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/palemoky/fight-the-landlord/internal/logger"
	"github.com/palemoky/fight-the-landlord/internal/ui"
	"github.com/palemoky/fight-the-landlord/internal/update"
)

// 默认服务器地址与版本号（可通过编译时 -ldflags 注入）
var (
	defaultServer = "localhost:1780"
	version       = "dev"
)

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
	showVersion := flag.Bool("version", false, "显示版本号并退出")
	skipUpdate := flag.Bool("no-update-check", false, "跳过启动时的更新检测")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ddz %s\n", version)
		return
	}

	// 支持完整 URL (wss://...) 或仅 host:port
	var serverURL string
	if strings.HasPrefix(*serverAddr, "ws://") || strings.HasPrefix(*serverAddr, "wss://") {
		serverURL = *serverAddr
	} else {
		serverURL = fmt.Sprintf("ws://%s/ws", *serverAddr)
	}

	// 启动前检测版本要求（失败不影响正常启动）
	if !*skipUpdate {
		checkForUpdate(serverURL)
	}

	logger.LogInfo("Connecting to server: %s", serverURL)

	model := ui.NewOnlineModel(serverURL)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		logger.LogError("Client error: %v", err)
		log.Printf("启动客户端时出错: %v", err)
	}
}

// checkForUpdate 由服务端驱动版本检测：向服务端查询其要求的最低客户端版本，仅当本地版本低于该最低版本时才强制升级。
// 这样升级策略由服务端集中控制——服务端只在确有不兼容变更时抬高最低版本，避免每次发版都打扰所有用户。开发版本（未注入版本号）跳过检测；查询失败（如无网络或服务端不支持该接口）仅记录日志，不阻断启动。
func checkForUpdate(serverURL string) {
	if update.IsDevVersion(version) {
		return
	}

	req, err := update.FetchServerRequirement(context.Background(), serverURL)
	if err != nil {
		logger.LogInfo("Server version requirement check skipped: %v", err)
		return
	}

	if !req.RequiresUpgrade(version) {
		return
	}

	logger.LogInfo("Server requires client >= %s (current %s), forcing upgrade", req.MinClientVersion, version)
	fmt.Printf("\n🚀 低于最低版本 %s（当前 %s），正在自动升级……\n", req.MinClientVersion, version)

	forceUpgrade(req.MinClientVersion)
}

// forceUpgrade 下载最新发布版本并替换当前可执行文件，随后以新版本重启。
//
// minVersion 仅用于失败时的提示信息。任何环节失败都会退出进程（强制升级语义下
// 不能带着不兼容的旧版本继续运行），并打印手动升级指引。
func forceUpgrade(minVersion string) {
	res, err := update.Check(context.Background(), version)
	if err != nil {
		failUpgrade(fmt.Sprintf("无法获取最新版本：%v", err), minVersion, "")
	}

	if err := update.Apply(context.Background(), res); err != nil {
		logger.LogError("Self-update failed: %v", err)
		failUpgrade(fmt.Sprintf("自动升级失败：%v", err), minVersion, res.ReleaseURL)
	}

	fmt.Printf("✅ 已升级到 %s，正在以新版本重新启动……\n\n", res.LatestVersion)
	if err := update.Relaunch(); err != nil {
		logger.LogError("Relaunch failed: %v", err)
		fmt.Fprintf(os.Stderr, "❌ 重新启动失败：%v\n", err)
		fmt.Fprintf(os.Stderr, "   升级已完成，请重新运行 ddz。\n")
		os.Exit(1)
	}
}

// failUpgrade 打印升级失败信息与手动升级指引并退出。
func failUpgrade(reason, minVersion, releaseURL string) {
	fmt.Fprintf(os.Stderr, "\n❌ %s\n", reason)
	fmt.Fprintf(os.Stderr, "   低于最低版本 %s，请手动升级后再试。\n", minVersion)
	if releaseURL != "" {
		fmt.Fprintf(os.Stderr, "   下载地址：%s\n", releaseURL)
	}
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

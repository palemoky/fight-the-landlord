package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/palemoky/fight-the-landlord/internal/config"
	"github.com/palemoky/fight-the-landlord/internal/server"
)

// version 是服务端版本号，可通过编译时 -ldflags "-X main.version=..." 注入。
var version = "dev"

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 将注入的版本号传递给 server 包，供 /version 接口公布
	server.Version = version

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("加载配置文件失败，使用默认配置: %v", err)
		cfg = config.Default()
	}

	// 创建服务器
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 监听关闭信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动服务器
	go func() {
		log.Println("🎮 斗地主服务器启动中...")
		if err := srv.Start(); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	<-ctx.Done()
	log.Println("📢 收到关闭信号，开始优雅关闭...")
	srv.GracefulShutdown(cfg.Game.ShutdownTimeoutDuration())
}

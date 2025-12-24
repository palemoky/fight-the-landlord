package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/palemoky/fight-the-landlord/internal/config"
	"github.com/palemoky/fight-the-landlord/internal/network/server"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

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

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("正在关闭服务器...")
		srv.Shutdown()
		os.Exit(0)
	}()

	// 启动服务器
	log.Println("🎮 斗地主服务器启动中...")
	if err := srv.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

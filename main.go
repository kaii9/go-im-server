package main

import (
	"fmt"

	"go-im-server/config"
	"go-im-server/db"
	"go-im-server/router"
	"go-im-server/ws"

	"go.uber.org/zap"
)

func main() {
	mainLog, _ := zap.NewProduction()
	defer mainLog.Sync()

	if err := config.Load(); err != nil {
		mainLog.Fatal("加载配置失败", zap.Error(err))
	}

	if err := db.Init(); err != nil {
		mainLog.Fatal("数据库连接失败", zap.Error(err))
	}
	mainLog.Info("数据库连接成功")

	go ws.DefaultHub.Run()

	r := router.SetupRouter()
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	mainLog.Info("服务启动", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		mainLog.Fatal("服务启动失败", zap.Error(err))
	}
}

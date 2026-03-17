package main

import (
	"GoLinko/internal/config"
	"GoLinko/internal/dao"
	httpserver "GoLinko/internal/http_server"
	"GoLinko/internal/service/chat"
	"fmt"
)

func main() {
	Setup()
}
func Setup() {
	fmt.Println("Setup.....")

	// 先加载配置
	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 自动迁移数据库表结构
	dao.DbAutoMigrate()
	chat.StartKafkaPipeline()

	// 配置加载后再初始化路由
	httpserver.InitRouter()

	// 启动 HTTP 服务器
	httpserver.StartHTTPServer()
}

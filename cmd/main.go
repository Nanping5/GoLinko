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

	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	dao.DbAutoMigrate()
	chat.StartKafkaPipeline()
	httpserver.InitRouter()
	httpserver.StartHTTPServer()
}

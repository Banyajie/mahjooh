package main

import (
	"chess_alg_jx/api"
	"chess_alg_jx/config"
	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/jx/room"
	"chess_alg_jx/logger"
	"chess_alg_jx/utils"
)

func init() {
	if err := config.InitConfig(); err != nil {
		panic(err)
	}
	logger.InitLog()
	if err := api.InitDB(config.Config.MysqlDb); err != nil {
		logger.Fatal(err)
	}
	if err := jx_nsq.InitNsq(); err != nil {
		logger.Fatal(err)
	}
	utils.InitRedis()
}

func main() {
	go room.GameNsqConsumer()
	api.RouteRegister()
}

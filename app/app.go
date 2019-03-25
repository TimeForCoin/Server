package app

import (
	"github.com/TimeForCoin/Server/app/configs"
	"github.com/TimeForCoin/Server/app/controllers"
	"github.com/TimeForCoin/Server/app/model"
	"github.com/kataras/iris"
)

// Run 程序入口
func Run(configPath string) {
	var config configs.Config
	config.GetConf(configPath)
	// 初始化数据库
	if err := model.InitDB(&config.Db); err != nil {
		panic(err)
	}
	// 初始化 Redis
	if err := model.InitRedis(&config.Redis); err != nil {
		panic(err)
	}
	// 启动服务器
	app := controllers.NewApp()

	if config.HTTP.Dev {
		app.Logger().SetLevel("debug")
	}

	if err := app.Run(iris.Addr(config.HTTP.Host + ":" + config.HTTP.Port)); err != nil {
		panic(err)
	}
}

package app

import (
	"github.com/TimeForCoin/Server/app/configs"
	"github.com/TimeForCoin/Server/app/controllers"
	"github.com/TimeForCoin/Server/app/model"
	"github.com/kataras/iris"
)

func Run(configPath string) {
	var config configs.Config
	config.GetConf(configPath)

	err := model.InitDB(config.Db)
	if err != nil {
		panic(err)
	}

	app := controllers.NewApp()

	if config.HTTP.Dev {
		app.Logger().SetLevel("debug")
	}

	if err := app.Run(iris.Addr(config.HTTP.Host + ":" + config.HTTP.Port)); err != nil {
		panic(err)
	}
}

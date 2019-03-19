package app

import (
	"github.com/TimeForCoin/Server/app/controllers"
	"github.com/kataras/iris"
)

func Run(configPath string) {
	var config Config
	config.GetConf(configPath)
	app := controllers.NewApp()

	if config.Server.Dev {
		app.Logger().SetLevel("debug")
	}

	if err := app.Run(iris.Addr(config.Server.Host + ":" + config.Server.Port)); err != nil {
		panic(err)
	}
}

package app

import (
	"github.com/TimeForCoin/Server/app/controllers"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/mvc"
)

func Run(configPath string) {
	var config Config
	config.GetConf(configPath)
	app := newApp()
	if err := app.Run(iris.Addr(":" + config.Server.Port)); err != nil {
		panic(err)
	}
}

func newApp() *iris.Application {
	app := iris.New()
	// recover from any http-relative panics
	app.Use(recover.New())
	// log the requests to the terminal.
	app.Use(logger.New())
	// Serve a controller based on the root Router, "/".
	mvc.New(app).Handle(new(controllers.UserController))
	return app
}

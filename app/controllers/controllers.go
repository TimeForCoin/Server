package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
)

// NewApp 创建服务器实例并绑定控制器
func NewApp() *iris.Application {
	app := iris.New()
	// recover from any http-relative panics
	app.Use(recover.New())
	// log the requests to the terminal.
	app.Use(logger.New())

	BindUserController(app)

	return app
}

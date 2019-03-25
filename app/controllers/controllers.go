package controllers

import (
	"time"

	"github.com/TimeForCoin/Server/app/configs"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
)

var sessionManager *sessions.Sessions

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

// InitSession 初始化 Session
func InitSession(config configs.SessionConfig) {
	sessionManager = sessions.New(sessions.Config{
		Cookie:  config.Key,
		Expires: time.Hour * time.Duration(config.Expires*24),
	})
}

func getSession() *sessions.Sessions {
	if sessionManager == nil {
		// 生成默认 Session
		sessionManager = sessions.New(sessions.Config{
			Cookie:  "coin-for-time",
			Expires: time.Hour * time.Duration(15*24),
		})
	}
	return sessionManager
}

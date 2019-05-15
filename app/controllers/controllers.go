package controllers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/sessions"
)

var sessionManager *sessions.Sessions

// NewApp 创建服务器实例并绑定控制器
func NewApp() *iris.Application {
	app := iris.New()
	// recover from any http-relative panics
	// log the requests to the terminal.
	app.Use(logger.New())

	app.Use(libs.NewErrorHandler())

	BindUserController(app)

	return app
}

type BaseController struct {
	Ctx     iris.Context
	Session *sessions.Session
}

// 检查登陆状态
func (b *BaseController) checkLogin() string {
	id := b.Session.GetString("id")
	_, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "invalid_session", 401)
	return id
}

func (b *BaseController) JSON(data interface{}) {
	libs.JSON(b.Ctx, data)
}

// InitSession 初始化 Session
func InitSession(config libs.SessionConfig) {
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

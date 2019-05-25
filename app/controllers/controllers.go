package controllers

import (
	"time"

	irisRecover "github.com/kataras/iris/middleware/recover"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

	app.Use(irisRecover.New())

	app.Use(libs.NewErrorHandler())

	BindUserController(app)

	return app
}

// BaseController 控制基类
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

// JSON 使用 JSON 返回数据
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

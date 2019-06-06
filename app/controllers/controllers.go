package controllers

import (
	"github.com/kataras/iris/sessions/sessiondb/redis"
	"github.com/kataras/iris/sessions/sessiondb/redis/service"
	"github.com/rs/zerolog/log"
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
	BindTaskController(app)
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
func InitSession(config libs.SessionConfig, dbConfig libs.RedisConfig) {
	sessionManager = sessions.New(sessions.Config{
		Cookie:  config.Key,
		Expires: time.Hour * time.Duration(config.Expires*24),
	})
	// 生成默认 Session
	db := redis.New(service.Config{
		Network:     "tcp",
		Addr:        dbConfig.Host + ":" + dbConfig.Port,
		Password:    dbConfig.Password,
		Database:    dbConfig.Session,
		MaxIdle:     0,
		MaxActive:   0,
		IdleTimeout: time.Duration(5) * time.Minute,
		Prefix:      "session"})

	// close connection when control+C/cmd+C
	iris.RegisterOnInterrupt(func() {
		if 	err := db.Close(); err != nil {
			log.Error().Msg(err.Error())
		}
	})
	sessionManager.UseDatabase(db)
}

func getSession() *sessions.Sessions {
	if sessionManager == nil {
		sessionManager = sessions.New(sessions.Config{
			Cookie:  "coin-for-time",
			Expires: time.Hour * time.Duration(15*24),
		})

		// sessionManager.UseDatabase()
	}
	return sessionManager
}

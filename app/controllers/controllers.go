package controllers

import (
	"strconv"
	"time"

	"github.com/TimeForCoin/Server/app/utils"

	"github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"github.com/rs/zerolog/log"

	irisRecover "github.com/kataras/iris/v12/middleware/recover"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/sessions"
)

var sessionManager *sessions.Sessions

// NewApp 创建服务器实例并绑定控制器
func NewApp() *iris.Application {
	app := iris.New()

	app.Use(logger.New())

	app.Use(irisRecover.New())

	app.Use(utils.NewErrorHandler())

	BindUserController(app)
	BindArticleController(app)
	BindTaskController(app)
	BindFileController(app)
	BindQuestionnaireController(app)
	BindCommentController(app)
	BindMessageController(app)
	BindUtilsController(app)

	return app
}

type omit *struct{}

// BaseController 控制基类
type BaseController struct {
	Ctx     iris.Context
	Session *sessions.Session
}

// PaginationRes 分页数据结构
type PaginationRes struct {
	Page  int64
	Size  int64
	Total int64
}

// 检查登陆状态
func (b *BaseController) checkLogin() primitive.ObjectID {
	id := b.Session.GetString("id")
	_id, err := primitive.ObjectIDFromHex(id)
	utils.Assert(err == nil, "invalid_session", 401)
	// login := b.Session.GetString("login")
	// libs.Assert(login != "wechat_new", "invalid_session", 401)
	return _id
}

func (b *BaseController) getPaginationData() (page, size int64) {
	var err error
	pageStr := b.Ctx.URLParamDefault("page", "1")
	page, err = strconv.ParseInt(pageStr, 10, 64)
	utils.AssertErr(err, "invalid_page", 400)
	utils.Assert(page > 0, "invalid_page", 400)
	sizeStr := b.Ctx.URLParamDefault("size", "10")
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	utils.AssertErr(err, "invalid_size", 400)
	utils.Assert(size > 0, "invalid_size", 400)
	return
}

// JSON 使用 JSON 返回数据
func (b *BaseController) JSON(data interface{}) {
	utils.JSON(b.Ctx, data)
}

// InitSession 初始化 Session
func InitSession(config utils.SessionConfig, dbConfig utils.RedisConfig) {
	sessionManager = sessions.New(sessions.Config{
		Cookie:  config.Key,
		Expires: time.Hour * time.Duration(config.Expires*24),
	})

	// 生成默认 Session
	db := redis.New(redis.Config{
		Network:   "tcp",
		Addr:      dbConfig.Host + ":" + dbConfig.Port,
		Timeout:   time.Duration(5) * time.Minute,
		MaxActive: 0,
		Password:  dbConfig.Password,
		Database:  dbConfig.Session,
		Prefix:    "session",
		Delim:     "-",
		Driver:    redis.Redigo(), // redis.Radix() can be used instead.
	})

	// close connection when control+C/cmd+C
	iris.RegisterOnInterrupt(func() {
		if err := db.Close(); err != nil {
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
	}
	return sessionManager
}

package app

import (
	"github.com/TimeForCoin/Server/app/controllers"
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/json-iterator/go/extra"
	"github.com/kataras/iris"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func initService(config libs.Config) {
	// 初始化数据库
	if err := models.InitDB(&config.Db); err != nil {
		panic(err)
	}
	// 初始化 Redis
	if err := models.InitRedis(&config.Redis); err != nil {
		panic(err)
	}
	// 初始化 Session
	controllers.InitSession(config.HTTP.Session)
	// 初始化 Json 设置
	// 自动转换成小写下划线风格
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	// 初始化 Violet Oauth 系统
	libs.InitViolet(config.Violet)
}

// Run 程序入口
func Run(configPath string) {
	// 初始化日志
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// 读取配置
	var config libs.Config
	config.LoadConf(configPath)
	// 初始化各种服务
	initService(config)
	// 启动服务器
	app := controllers.NewApp()

	if config.Dev {
		app.Logger().SetLevel("debug")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	if err := app.Run(iris.Addr(config.HTTP.Host + ":" + config.HTTP.Port)); err != nil {
		panic(err)
	}
}

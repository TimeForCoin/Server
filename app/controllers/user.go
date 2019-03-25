package controllers

import (
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"
)

// UserController 用户控制
type UserController struct {
	Ctx iris.Context
	Server services.UserService
	Session *sessions.Session
}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	userRoute := mvc.New(app.Party("/user"))
	userRoute.Register(services.NewUserService(), getSession().Start)
	userRoute.Handle(new(UserController))
}

// GetPing 测试用
// Method:   GET
// Resource: http://localhost:port/user/ping
func (c *UserController) GetPing() string {
	res := c.Server.GetPong("ping")
	return res
}

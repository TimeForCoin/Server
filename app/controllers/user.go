package controllers

import (
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

// UserController 用户控制
type UserController struct{
	Server services.UserService
}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	userRoute :=  mvc.New(app.Party("/user"))
	userRoute.Register(services.NewUserService())
	userRoute.Handle(new(UserController))
}

// GetPing 测试用
// Method:   GET
// Resource: http://localhost:port/user/ping
func (c *UserController) GetPing() string {
	res := c.Server.GetPong("ping")
	return res
}

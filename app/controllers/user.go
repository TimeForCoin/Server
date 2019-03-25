package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

// UserController 用户控制
type UserController struct{}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	mvc.New(app.Party("/user")).Handle(new(UserController))
}

// GetPing 测试用
// Method:   GET
// Resource: http://localhost:port/user/ping
func (c *UserController) GetPing() string {
	return "pong"
}

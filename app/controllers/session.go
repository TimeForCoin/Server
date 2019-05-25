package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
)

// SessionController 用户登陆状态控制
type SessionController struct {
	BaseController
	Server services.UserService
}

// GetSessionRes 获取登陆URL返回值
type GetSessionRes struct {
	URL string `json:"url"`
}

// Get 获取 Violet 登陆授权 URL
func (c *SessionController) Get() int {
	loginURL, state := c.Server.GetLoginURL()
	c.Session.Set("state", state)
	c.Session.Set("login", "running")
	libs.JSON(c.Ctx, GetSessionRes{
		URL: loginURL,
	})
	return iris.StatusOK
}

// GetVioletReq Violet 登陆参数
type GetVioletReq struct {
	Code  string
	State string
}

// GetViolet 通过 Violet 登陆
func (c *SessionController) GetViolet() int {
	defer func() {
		if err := recover(); err != nil {
			c.Session.Set("login", "false")
			panic(err)
		}
	}()
	code := c.Ctx.FormValue("code")
	state := c.Ctx.FormValue("state")
	libs.Assert(code != "" && state != "", "登陆失败，请重试")

	rightState := c.Session.GetString("state")
	libs.Assert(state == rightState, "状态校验失败，请重试")
	c.Session.Delete("state")

	id, _ := c.Server.LoginByCode(code)
	libs.Assert(id != "", "登陆已过期，请重试")

	c.Session.Set("id", id)
	c.Session.Set("login", "violet")
	return iris.StatusCreated
}

// GetSessionStatusRes 获取登陆状态返回值
type GetSessionStatusRes struct {
	Status string
}

// GetStatus 获取登陆状态
func (c *SessionController) GetStatus() int {

	status := c.Session.GetString("login")
	if status == "" {
		status = "none"
	}
	c.JSON(GetSessionStatusRes{
		Status: status,
	})
	return iris.StatusOK
}

// Delete 退出登陆
func (c *SessionController) Delete() int {
	c.Session.Set("login", "none")
	c.Session.Delete("id")
	return iris.StatusOK
}

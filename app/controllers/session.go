package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
)

// SessionController 用户登陆状态控制
type SessionController struct {
	BaseController
	Service services.UserService
}

// GetSessionRes 获取登陆URL返回值
type GetSessionRes struct {
	URL string `json:"url"`
}

// Get 获取 Violet 登陆授权 URL
func (c *SessionController) Get() int {
	loginURL, state := c.Service.GetLoginURL()
	c.Session.Set("state", state)
	c.Session.Set("login", "running")
	libs.JSON(c.Ctx, GetSessionRes{
		URL: loginURL,
	})
	return iris.StatusOK
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

	id, _ := c.Service.LoginByViolet(code)
	libs.Assert(id != "", "登陆已过期，请重试")

	c.Session.Set("id", id)
	c.Session.Set("login", "violet")
	return iris.StatusCreated
}

// GetVioletReq Violet 登陆参数
type PostWechatRes struct {
	New bool
}

// GetVioletReq Violet 登陆参数
type PostWechatReq struct {
	Code string
}

// GetWechat 通过微信登陆
func (c *SessionController) PostWechat() int {
	req := PostWechatReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil && req.Code != "", "invalid_code", 400)

	id, newUser := c.Service.LoginByWechat(req.Code)
	c.JSON(PostWechatRes{
		New: newUser,
	})
	c.Session.Set("id", id)
	if newUser {
		c.Session.Set("login", "wechat_new")
	} else {
		c.Session.Set("login", "wechat")
	}
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

// 测试
func (c *SessionController) GetTest() string {
	return "Test"
}

type GetWeChatImageRes struct {
	Data string
}

// 微信登陆二维码
func (c *SessionController) GetWechat() int {
	// id := c.checkLogin()
	image, err := libs.GetWechat().MakeImage("hello")
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	c.JSON(GetWeChatImageRes{
		Data: image,
	})
	return iris.StatusOK
}
package controllers

import (
	"github.com/TimeForCoin/Server/app/services"
	"github.com/TimeForCoin/Server/app/utils"
	"github.com/kataras/iris/v12"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	utils.JSON(c.Ctx, GetSessionRes{
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
	utils.Assert(code != "" && state != "", "登陆失败，请重试")

	rightState := c.Session.GetString("state")
	utils.Assert(state == rightState, "状态校验失败，请重试")
	c.Session.Delete("state")

	id, _ := c.Service.LoginByViolet(code)
	utils.Assert(id != "", "登陆已过期，请重试")

	c.Session.Set("id", id)
	c.Session.Set("login", "violet")
	return iris.StatusOK
}

// PostWechatRes 微信登陆数据
type PostWechatRes struct {
	New bool
}

// PostWechatReq 温馨登陆请求
type PostWechatReq struct {
	Code string
}

// PostWechat 通过微信登陆
func (c *SessionController) PostWechat() int {
	req := PostWechatReq{}
	err := c.Ctx.ReadJSON(&req)
	utils.Assert(err == nil && req.Code != "", "invalid_code", 400)

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
	return iris.StatusOK
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
	if status == "wechat_qr" {
		session := c.Session.GetString("qr-code")
		sessionID, err := primitive.ObjectIDFromHex(session)
		if err != nil {
			status = "none"
			c.Session.Set("login", status)
		} else {
			id := c.Service.GetSessionUser(sessionID)
			if id != primitive.NilObjectID {
				c.Session.Set("id", id.Hex())
				status = "wechat_pc"
				c.Session.Set("login", status)
			}
		}
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

// GetWeChatImageRes 获取微信登陆位置码
type GetWeChatImageRes struct {
	Data string
}

// GetWechat 获取微信登陆二维码
func (c *SessionController) GetWechat() int {
	sessionID := primitive.NewObjectID()
	png, err := qrcode.Encode(sessionID.Hex(), qrcode.Medium, 256)
	c.Session.Set("qr-code", sessionID.Hex())
	c.Session.Set("login", "wechat_qr")
	utils.AssertErr(err, "", 500)
	_, err = c.Ctx.Write(png)
	utils.AssertErr(err, "", 500)
	return iris.StatusOK
}

type PutWechatReq struct {
	Session string
}

// PutWechat 微信扫码登陆
func (c *SessionController) PutWechat() int {
	userID := c.checkLogin()
	req := PutWechatReq{}
	err := c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)
	sessionID, err := primitive.ObjectIDFromHex(req.Session)
	utils.AssertErr(err, "invalid_session", 400)
	c.Service.LoginByWechatOnPC(userID, sessionID)
	return iris.StatusOK
}

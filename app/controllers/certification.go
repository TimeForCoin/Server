package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserController 用户控制
type CertificationController struct {
	BaseController
	Server services.UserService
}

type PostCertificationReq struct {
	Identity models.UserIdentity
	Data string
	Type string
	Attachment []string
	Email string
}

// 用户认证
func (c *CertificationController) Post() int {
	id := c.checkLogin()
	req := PostCertificationReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_valid", 400)
	// 目前只支持学生认证
	libs.Assert(req.Identity == models.IdentityStudent, "invalid_identity", 400)
	// 目前只支持邮箱认证
	libs.Assert(req.Type == "email", "invalid_type", 400)
	libs.Assert(libs.IsEmail(req.Email), "invalid_email", 400)

	c.Server.AddEmailCertification(req.Identity, id, req.Data, req.Email)
	return iris.StatusOK
}

func (c *CertificationController) PostEmail() int {
	id := c.checkLogin()
	c.Server.SendCertificationEmail(id, "")
	return iris.StatusOK
}

func (c *CertificationController) GetAuth() string {
	code := c.Ctx.FormValue("code")
	user := c.Ctx.FormValue("user")
	if code == "" || user == "" {
		return "无效的认证链接"
	}
	return c.Server.CheckCertification(user, code)
}

type PutUserReq struct {
	Operate string
	Data string
	Feedback string
}

// 更新认证
func (c *CertificationController) PutUserBy(id string) int {
	sessionUser := c.checkLogin()
	if id == "me" {
		id = sessionUser
	}
	_, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "invalid_id")
	req := PutUserReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	if req.Operate == "cancel" {
		c.Server.CancelCertification(id)
	} else if req.Operate == "true" || req.Operate == "false" {
		// 验证管理员权限
		user, err := models.GetRedis().Cache.GetUserBaseInfo(sessionUser)
		libs.AssertErr(err, "invalid_session", 401)
		libs.Assert(user.Type == models.UserTypeRoot || user.Type == models.UserTypeAdmin, "permission_deny", 403)
		if req.Operate == "true" {
			c.Server.UpdateCertification(id, req.Operate, req.Data)
		} else {
			c.Server.UpdateCertification(id, req.Operate, req.Feedback)
		}
	} else {
		libs.Assert(false, "invalid_operate", 400)
	}
	return iris.StatusOK
}
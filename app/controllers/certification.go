package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CertificationController 认证API
type CertificationController struct {
	BaseController
	Service services.UserService
}

// PostCertificationReq 申请认证请求
type PostCertificationReq struct {
	Identity   models.UserIdentity
	Data       string
	Type       string
	Attachment []string
	Email      string
}

// Post 提交用户认证
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
	// TODO 材料认证
	c.Service.AddEmailCertification(req.Identity, id, req.Data, req.Email)
	return iris.StatusOK
}

// PostEmail 再次发送认证邮件
func (c *CertificationController) PostEmail() int {
	id := c.checkLogin()
	c.Service.SendCertificationEmail(id, "")
	return iris.StatusOK
}

// GetAuth 验证认证链接
func (c *CertificationController) GetAuth() string {
	code := c.Ctx.FormValue("code")
	user := c.Ctx.FormValue("user")

	if code == "" || user == "" {
		return "无效的认证链接"
	}
	userID, err := primitive.ObjectIDFromHex(user)
	if err != nil {
		return "无效的认证链接"
	}
	return c.Service.CheckCertification(userID, code)
}

// PutUserReq 更新认证请求
type PutUserReq struct {
	Operate  string
	Data     string
	Feedback string
}

// PutUserBy 更新认证
func (c *CertificationController) PutUserBy(userID string) int {
	sessionUser := c.checkLogin()
	var opUser primitive.ObjectID
	if userID == "me" {
		opUser = sessionUser
	} else {
		var err error
		opUser, err = primitive.ObjectIDFromHex(userID)
		libs.AssertErr(err, "invalid_id", 400)
	}
	req := PutUserReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	if req.Operate == "cancel" {
		c.Service.CancelCertification(opUser)
	} else if req.Operate == "true" || req.Operate == "false" {
		// 验证管理员权限
		user, err := models.GetRedis().Cache.GetUserBaseInfo(sessionUser)
		libs.AssertErr(err, "invalid_session", 401)
		libs.Assert(user.Type == models.UserTypeRoot || user.Type == models.UserTypeAdmin, "permission_deny", 403)
		if req.Operate == "true" {
			c.Service.UpdateCertification(opUser, req.Operate, req.Data)
		} else {
			c.Service.UpdateCertification(opUser, req.Operate, req.Feedback)
		}
	} else {
		libs.Assert(false, "invalid_operate", 400)
	}
	return iris.StatusOK
}

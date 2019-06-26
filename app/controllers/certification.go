package controllers

import (
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/TimeForCoin/Server/app/utils"
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
	utils.AssertErr(err, "invalid_valid", 400)
	// 目前只支持学生认证
	utils.Assert(req.Identity == models.IdentityStudent, "invalid_identity", 400)
	// 目前只支持邮箱认证
	switch req.Type {
	case "email":
		utils.Assert(utils.IsEmail(req.Email), "invalid_email", 400)
		c.Service.AddEmailCertification(req.Identity, id, req.Data, req.Email)
	case "material":
		var material []primitive.ObjectID
		for _, attachment := range req.Attachment {
			fileID, err := primitive.ObjectIDFromHex(attachment)
			utils.AssertErr(err, "invalid_attachment", 400)
			material = append(material, fileID)
		}
		utils.Assert(len(material) > 0, "invalid_attachment", 400)
		c.Service.AddMaterialCertification(req.Identity, id, req.Data, material)
	default:
		utils.Assert(false, "invalid_type", 400)
	}
	return iris.StatusOK
}

// PostEmail 再次发送认证邮件
func (c *CertificationController) PostEmail() int {
	id := c.checkLogin()
	c.Service.SendCertificationEmail(id, "")
	return iris.StatusOK
}

// GetCertification 获取待审核认证列表
func (c *CertificationController) GetCertification() int {
	userID := c.checkLogin()
	page, size := c.getPaginationData()
	types := c.Ctx.URLParamDefault("type", "email")
	var status []models.CertificationStatus
	switch types {
	case "email":
		status = append(status, models.CertificationWaitEmail)
	case "material":
		status = append(status, models.CertificationWait)
	case "all":
		status = append(status, models.CertificationWaitEmail)
		status = append(status, models.CertificationWait)
	default:
		utils.Assert(false, "invalid_type", 400)
	}
	users := c.Service.GetCertificationList(userID, status, page, size)

	c.JSON(struct {
		Pagination PaginationRes
		Data       []services.UserDetail
	}{
		Pagination: PaginationRes{
			Page: page,
			Size: size,
		},
		Data: users,
	})

	return iris.StatusOK
}

// GetAuto 获取自动认证前缀
func (c *CertificationController) GetAuto() int {
	userID := c.checkLogin()
	page, size := c.getPaginationData()
	data := c.Service.GetAutoCertification(userID, page, size)
	c.JSON(struct {
		Pagination PaginationRes
		Data       []models.SystemSchemas
	}{
		Pagination: PaginationRes{
			Page: page,
			Size: size,
		},
		Data: data,
	})
	return iris.StatusOK
}

// PostAuto 添加自动认证前缀
func (c *CertificationController) PostAuto() int {
	userID := c.checkLogin()
	req := struct {
		Key   string
		Value string
	}{}
	err := c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)
	utils.Assert(req.Key != "" && req.Value != "", "invalid_value", 400)
	c.Service.AddAutoCertification(userID, req.Key, req.Value)
	return iris.StatusOK
}

// DeleteAutoBy 删除自动认证前缀
func (c *CertificationController) DeleteAutoBy(key string) int {
	userID := c.checkLogin()
	utils.Assert(key != "", "invalid_key", 400)
	c.Service.RemoveAutoCertification(userID, key)
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
		utils.AssertErr(err, "invalid_id", 400)
	}
	req := PutUserReq{}
	err := c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)
	if req.Operate == "cancel" {
		c.Service.CancelCertification(opUser)
	} else if req.Operate == "true" || req.Operate == "false" {
		// 验证管理员权限
		user, err := models.GetRedis().Cache.GetUserBaseInfo(sessionUser)
		utils.AssertErr(err, "invalid_session", 401)
		utils.Assert(user.Type == models.UserTypeRoot || user.Type == models.UserTypeAdmin, "permission_deny", 403)
		if req.Operate == "true" {
			c.Service.UpdateCertification(opUser, req.Operate, req.Data)
		} else {
			c.Service.UpdateCertification(opUser, req.Operate, req.Feedback)
		}
	} else {
		utils.Assert(false, "invalid_operate", 400)
	}
	return iris.StatusOK
}

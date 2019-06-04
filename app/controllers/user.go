package controllers

import (
	"reflect"
	"strings"
	"time"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserController 用户控制
type UserController struct {
	BaseController
	Server services.UserService
}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	userService := services.NewUserService()

	sessionRoute := mvc.New(app.Party("/session"))
	sessionRoute.Register(userService, getSession().Start)
	sessionRoute.Handle(new(SessionController))

	userRoute := mvc.New(app.Party("/users"))
	userRoute.Register(userService, getSession().Start)
	userRoute.Handle(new(UserController))
}

// GetInfoByIDRes 获取用户信息返回值
type GetInfoByIDRes struct {
	ID           string `json:"id"`
	VioletName   string `json:"violet_name,omitempty"`
	WechatName   string `json:"wechat_name,omitempty"`
	RegisterTime int64
	Info         models.UserInfoSchema
	Data         *UserDataRes
	Certification *UserCertification
}


type omit *struct{}
// UserDataRes 用户数据返回值
type UserDataRes struct {
	*models.UserDataSchema
	// 额外项
	Attendance bool
	// 排除项
	AttendanceDate omit `json:"attendance_date,omitempty"`
	CollectTasks   omit `json:"collect_tasks,omitempty"`
	SearchHistory  omit `json:"search_history,omitempty"`
}

// UserCertification 用户认证信息
type UserCertification struct {
	Type models.UserIdentity
	Data string
	Date int64
}

// GetInfoBy 获取用户信息
func (c *UserController) GetInfoBy(id string) int {
	if id == "me" {
		id = c.checkLogin()
	}
	_, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "invalid_id")
	user := c.Server.GetUser(id)

	attendanceTime := time.Unix(user.Data.AttendanceDate, 0)
	nowTime := time.Now()
	isAttendance := attendanceTime.Year() == nowTime.Year() && attendanceTime.YearDay() == nowTime.YearDay()
	c.JSON(GetInfoByIDRes{
		ID:           user.ID.Hex(),
		VioletName:   user.VioletName,
		WechatName:   user.WechatName,
		RegisterTime: user.RegisterTime,
		Info:         user.Info,
		Data: &UserDataRes{
			UserDataSchema: &user.Data,
			Attendance:     isAttendance,
		},
		Certification: &UserCertification{
			Type: user.Certification.Identity,
			Data: user.Certification.Data,
			Date: user.Certification.Date,
		},
	})
	return iris.StatusOK
}

// PostAttend 用户签到
func (c *UserController) PostAttend() int {
	id := c.checkLogin()
	c.Server.UserAttend(id)
	return iris.StatusOK
}

type UserInfoReq struct {
	*models.UserInfoSchema
	AvatarURL string `json:"avatarUrl"`
}

// PatchInfo 修改用户信息
func (c *UserController) PutInfo() int {
	id := c.checkLogin()
	// 解析
	req := UserInfoReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	libs.Assert(req.Email == "" || libs.IsEmail(req.Email), "invalid_email", 400)
	libs.Assert(req.Gender == "" || libs.IsGender(string(req.Gender)), "invalid_gender", 400)
	// TODO 获取头像链接保存在本地存储库中
	if req.AvatarURL != "" {
		req.Avatar = req.AvatarURL
	} else if req.Avatar != "" {
		libs.Assert(strings.HasPrefix(req.Avatar,"data:image/png;base64,"), "invalid_avatar", 400)
		url, err := libs.GetCOS().SaveBase64("avatar-" + id + ".png", req.Avatar[len("data:image/png;base64,"):])
		libs.AssertErr(err, "invalid_avatar_decode", 400)
		req.Avatar = url
	}

	// 判断是否存在数据
	count := 0
	names := reflect.TypeOf(req)
	values := reflect.ValueOf(req)
	for i := 0; i < values.NumField(); i++ {
		name := names.Field(i).Tag.Get("bson")
		if name == "birthday" { // 生日字段为 int64
			if values.Field(i).Int() != 0 {
				count++
			}
		} else { // 其他字段为 string
			if values.Field(i).String() != "" {
				count++
			}
		}
	}
	libs.Assert(count != 0, "invalid_value", 400)

	c.Server.SetUserInfo(id, *req.UserInfoSchema)

	if c.Session.GetString("status") == "wechat_new" {
		c.Session.Set("status", "wechat")
	}

	return iris.StatusOK
}

// UserDataRes 用户数据返回值
type UseTypeReq struct {
	ID string `json:"id"`
	Type string `json:"type"`
}

// 修改用户信息
func (c *UserController) PutType() int {
	id := c.checkLogin()
	// 解析
	req := UseTypeReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	libs.Assert(libs.IsID(req.ID), "invalid_id", 400)
	libs.Assert(libs.IsUserType(req.Type) , "invalid_type", 400)
	libs.Assert(req.Type != string(models.UserTypeRoot), "not_allow_type", 403)
	c.Server.SetUserType(id, req.ID, models.UserType(req.Type))

	return iris.StatusOK
}
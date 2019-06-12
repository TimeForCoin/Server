package controllers

import (
	"reflect"
	"strings"
	"time"

	"strconv"

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
	Service services.UserService
}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	userService := services.GetServiceManger().User

	sessionRoute := mvc.New(app.Party("/session"))
	sessionRoute.Register(userService, getSession().Start)
	sessionRoute.Handle(new(SessionController))

	userRoute := mvc.New(app.Party("/users"))
	userRoute.Register(userService, getSession().Start)
	userRoute.Handle(new(UserController))

	certificationRoute := mvc.New(app.Party("/certification"))
	certificationRoute.Register(userService, getSession().Start)
	certificationRoute.Handle(new(CertificationController))
}

// GetInfoByIDRes 获取用户信息返回值
type GetInfoByIDRes struct {
	ID            string `json:"id"`
	VioletName    string `json:"violet_name,omitempty"`
	WechatName    string `json:"wechat_name,omitempty"`
	RegisterTime  int64
	Info          models.UserInfoSchema
	Data          *UserDataRes
	Certification *UserCertification
}

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
	Type   models.UserIdentity
	Status models.CertificationStatus
	Email  string
	Data   string
	Date   int64
}

// GetInfoBy 获取用户信息
func (c *UserController) GetInfoBy(userID string) int {
	var id primitive.ObjectID
	if userID == "me" {
		id = c.checkLogin()
	} else {
		var err error
		id, err = primitive.ObjectIDFromHex(userID)
		libs.AssertErr(err, "invalid_session", 401)
	}
	user := c.Service.GetUser(id)
	res := GetInfoByIDRes{
		ID:           user.ID.Hex(),
		VioletName:   user.VioletName,
		WechatName:   user.WechatName,
		RegisterTime: user.RegisterTime,
		Info:         user.Info,
		Data: &UserDataRes{
			UserDataSchema: &user.Data,
		},
		Certification: &UserCertification{
			Type:   user.Certification.Identity,
			Status: user.Certification.Status,
			Email:  user.Certification.Email,
			Data:   user.Certification.Data,
			Date:   user.Certification.Date,
		},
	}
	nowTime := time.Now()
	attendanceTime := time.Unix(user.Data.AttendanceDate, 0)
	res.Data.Attendance = attendanceTime.Year() == nowTime.Year() && attendanceTime.YearDay() == nowTime.YearDay()
	if userID != "me" {
		res.Certification.Email = ""
		if res.Certification.Status != models.CertificationTrue {
			res.Certification = &UserCertification{
				Type: models.IdentityNone,
			}
		}
	}
	c.JSON(res)
	return iris.StatusOK
}

// PostAttend 用户签到
func (c *UserController) PostAttend() int {
	id := c.checkLogin()
	c.Service.UserAttend(id)
	return iris.StatusOK
}

// PutUserInfoReq 修改用户信息请求
type PutUserInfoReq struct {
	*models.UserInfoSchema
	AvatarURL string `json:"avatar_url"`
}

// PutInfo 修改用户信息
func (c *UserController) PutInfo() int {
	id, err := primitive.ObjectIDFromHex(c.Session.GetString("id"))
	libs.Assert(err == nil, "invalid_session", 401)
	// 解析
	req := PutUserInfoReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	libs.Assert(req.Email == "" || libs.IsEmail(req.Email), "invalid_email", 400)
	libs.Assert(req.Gender == "" || libs.IsGender(string(req.Gender)), "invalid_gender", 400)

	// 处理头像数据
	if req.AvatarURL != "" {
		url, err := libs.GetCOS().SaveURLFile("avatar-"+id.Hex()+".png", req.AvatarURL)
		libs.AssertErr(err, "", 400)
		req.Avatar = url
	} else if req.Avatar != "" {
		libs.Assert(strings.HasPrefix(req.Avatar, "data:image/png;base64,"), "invalid_avatar", 400)
		url, err := libs.GetCOS().SaveBase64File("avatar-"+id.Hex()+".png", req.Avatar[len("data:image/png;base64,"):])
		libs.AssertErr(err, "", 400)
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

	c.Service.SetUserInfo(id, *req.UserInfoSchema)

	if c.Session.GetString("login") == "wechat_new" {
		c.Session.Set("login", "wechat")
	}

	return iris.StatusOK
}

// PutUserTypeReq 修改用户类型请求
type PutUserTypeReq struct {
	Type string `json:"type"`
}

// PutTypeByID 修改用户类型
func (c *UserController) PutTypeByID(userID string) int {
	id := c.checkLogin()
	// 解析
	req := PutUserTypeReq{}
	err := c.Ctx.ReadJSON(&req)

	var opID primitive.ObjectID
	if userID == "me" {
		opID = id
	} else {
		var err error
		opID, err = primitive.ObjectIDFromHex(userID)
		libs.AssertErr(err, "invalid_id", 400)
	}

	libs.Assert(err == nil, "invalid_value", 400)
	libs.Assert(libs.IsUserType(req.Type), "invalid_type", 400)
	libs.Assert(req.Type != string(models.UserTypeRoot), "not_allow_type", 403)
	c.Service.SetUserType(id, opID, models.UserType(req.Type))
	return iris.StatusOK
}

// GetCollect 获取用户收藏
func (c *UserController) GetCollect() int {
	pageStr := c.Ctx.URLParamDefault("page", "1")
	page, err := strconv.ParseInt(pageStr, 10, 64)
	libs.AssertErr(err, "invalid_page", 400)
	sizeStr := c.Ctx.URLParamDefault("size", "10")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	libs.AssertErr(err, "invalid_size", 400)
	userIDString := c.Ctx.URLParamDefault("user_id", "")
	libs.Assert(userIDString != "", "string")
	var userID primitive.ObjectID
	if userIDString == "me" {
		userID = c.checkLogin()
	} else {
		userID, err = primitive.ObjectIDFromHex(userIDString)
		libs.AssertErr(err, "invalid_user", 403)
	}

	sort := c.Ctx.URLParamDefault("sort", "new")
	taskType := c.Ctx.URLParamDefault("type", "all")
	status := c.Ctx.URLParamDefault("status", "wait")
	reward := c.Ctx.URLParamDefault("reward", "all")

	taskCount, tasksData := c.Service.GetUserCollections(userID, page, size, sort,
		taskType, status, reward)
	if tasksData == nil {
		tasksData = []services.TaskDetail{}
	}

	res := TasksListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: taskCount,
		},
		Tasks: tasksData,
	}
	c.JSON(res)
	return iris.StatusOK
}

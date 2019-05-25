package controllers

import (
	"reflect"
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

	userRoute := mvc.New(app.Party("/user"))
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
	})
	return iris.StatusOK
}

// PostAttend 用户签到
func (c *UserController) PostAttend() int {
	id := c.checkLogin()
	c.Server.UserAttend(id)
	return iris.StatusOK
}

// PatchInfo 修改用户信息
func (c *UserController) PatchInfo() int {
	id := c.checkLogin()
	// 解析
	req := models.UserInfoSchema{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	libs.Assert(req.Email == "" || libs.IsEmail(req.Email), "invalid_email", 400)
	libs.Assert(req.Gender == "" || libs.IsGender(string(req.Gender)), "invalid_gender", 400)
	// TODO 处理头像
	// 暂时不支持修改头像
	req.Avatar = ""
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

	c.Server.SetUserInfo(id, req)
	return iris.StatusOK
}

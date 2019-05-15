package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// UserController 用户控制
type UserController struct {
	BaseController
	Server  services.UserService
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

type GetInfoByIDRes struct {
	ID         string `json:"id"`
	VioletName string
	WechatName string
	RegisterTime int64
	Info       models.UserInfoSchema
	Data       UserDataRes
}

type UserDataRes struct {
	*models.UserDataSchema
	// 额外项
	Attendance bool
	// 排除项
	AttendanceDate int64                `json:"-"`
	CollectTasks   []primitive.ObjectID `json:"-"`
	SearchHistory  []string             `json:"-"`
}

func (c *UserController) GetInfoBy(id string) int {
	if id == "me" {
		id = c.checkLogin()
	}
	_, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "invalid_id")
	user, err := c.Server.GetUser(id)
	libs.Assert(err == nil, "faked_user")

	attendanceTime := time.Unix(user.Data.AttendanceDate, 0)
	nowTime := time.Now()
	isAttendance := attendanceTime.Year() == nowTime.Year() && attendanceTime.YearDay() == nowTime.YearDay()
	c.JSON(GetInfoByIDRes{
		ID: user.ID.Hex(),
		VioletName: user.VioletName,
		WechatName: user.WechatName,
		RegisterTime: user.RegisterTime,
		Info: user.Info,
		Data:UserDataRes{
			UserDataSchema: &user.Data,
			Attendance: isAttendance,
		},
	})
	return iris.StatusOK
}

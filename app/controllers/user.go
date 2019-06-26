package controllers

import (
	"github.com/TimeForCoin/Server/app/utils"
	"reflect"
	"strings"

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

// UserListRes 用户数据
type UserListRes struct {
	Pagination PaginationRes
	Data       []services.UserDetail
}

// Get 搜索用户
func (c *UserController) Get() int {
	page, size := c.getPaginationData()
	key := c.Ctx.URLParamDefault("key", "")
	utils.Assert(key != "", "invalid_key", 400)

	res := c.Service.SearchUser(key, page, size)

	// noinspection GoPreferNilSlice
	for i := range res {
		sessionUserID := c.Session.GetString("id")
		if sessionUserID != "" {
			sessionUser, err := primitive.ObjectIDFromHex(sessionUserID)
			utils.AssertErr(err, "", 500)
			res[i].Data.Follower = c.Service.IsFollower(sessionUser, res[i].UserID)
			res[i].Data.Following = c.Service.IsFollowing(sessionUser, res[i].UserID)
		}
	}

	c.JSON(UserListRes{
		Pagination: PaginationRes{
			Page: page,
			Size: size,
		},
		Data: res,
	})
	return iris.StatusOK
}

// GetInfoBy 获取用户信息
func (c *UserController) GetInfoBy(userID string) int {
	var id primitive.ObjectID
	if userID == "me" {
		id = c.checkLogin()
	} else {
		var err error
		id, err = primitive.ObjectIDFromHex(userID)
		utils.AssertErr(err, "invalid_session", 401)
	}
	res := c.Service.GetUser(id, userID == "me")
	sessionUserID := c.Session.GetString("id")
	if userID != "me" && sessionUserID != "" {
		sessionUser, err := primitive.ObjectIDFromHex(sessionUserID)
		utils.AssertErr(err, "", 500)
		res.Data.Follower = c.Service.IsFollower(sessionUser, res.UserID)
		res.Data.Following = c.Service.IsFollowing(sessionUser, res.UserID)
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

func (c *UserController) PostPay() int {
	c.Service.UserPay(c.checkLogin())
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
	utils.Assert(err == nil, "invalid_session", 401)
	// 解析
	req := PutUserInfoReq{}
	err = c.Ctx.ReadJSON(&req)
	utils.Assert(err == nil, "invalid_value", 400)
	utils.Assert(req.Email == "" || utils.IsEmail(req.Email), "invalid_email", 400)
	utils.Assert(req.Gender == "" || utils.IsGender(string(req.Gender)), "invalid_gender", 400)

	// 处理头像数据
	if req.AvatarURL != "" {
		url, err := libs.GetCOS().SaveURLFile("avatar-"+id.Hex()+".png", req.AvatarURL)
		utils.AssertErr(err, "", 400)
		req.Avatar = url
	} else if req.Avatar != "" {
		var url string
		if strings.HasPrefix(req.Avatar, "data:image/png;base64,") {
			url, err = libs.GetCOS().SaveBase64File("avatar-"+id.Hex()+".png", req.Avatar[len("data:image/png;base64,"):])
		} else if strings.HasPrefix(req.Avatar, "data:image/jpeg;base64,") {
			url, err = libs.GetCOS().SaveBase64File("avatar-"+id.Hex()+".jpg", req.Avatar[len("data:image/jpeg;base64,"):])
		}
		utils.AssertErr(err, "", 400)
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
	utils.Assert(count != 0, "invalid_value", 400)

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
		utils.AssertErr(err, "invalid_id", 400)
	}

	utils.Assert(err == nil, "invalid_value", 400)
	utils.Assert(utils.IsUserType(req.Type), "invalid_type", 400)
	utils.Assert(req.Type != string(models.UserTypeRoot), "not_allow_type", 403)
	c.Service.SetUserType(id, opID, models.UserType(req.Type))
	return iris.StatusOK
}

// GetCollectBy 获取用户收藏
func (c *UserController) GetCollectBy(userIDString string) int {
	page, size := c.getPaginationData()
	utils.Assert(userIDString != "", "string")
	var userID primitive.ObjectID
	if userIDString == "me" {
		userID = c.checkLogin()
	} else {
		var err error
		userID, err = primitive.ObjectIDFromHex(userIDString)
		utils.AssertErr(err, "invalid_user", 403)
	}

	sort := c.Ctx.URLParamDefault("sort", "new")
	taskType := c.Ctx.URLParamDefault("type", "all")
	status := c.Ctx.URLParamDefault("status", "wait")
	reward := c.Ctx.URLParamDefault("reward", "all")

	taskCount, tasksData := c.Service.GetUserCollections(userID, page, size, sort,
		taskType, status, reward)

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

// TasksStatusListRes 用户参与任务数据
type TasksStatusListRes struct {
	Pagination PaginationRes
	Data       []services.TaskStatusDetail
}

// GetTaskBy 获取用户参与的任务列表
func (c *UserController) GetTaskBy(userIDString string) int {
	page, size := c.getPaginationData()
	utils.Assert(userIDString != "", "string")
	var userID primitive.ObjectID
	var err error
	if userIDString == "me" {
		userID = c.checkLogin()
	} else {
		userID, err = primitive.ObjectIDFromHex(userIDString)
		utils.AssertErr(err, "invalid_user", 403)
	}

	status := c.Ctx.URLParamDefault("status", "all")

	taskCount, taskStatusesData := c.Service.GetUserParticipate(userID, page, size, status)
	if taskStatusesData == nil {
		taskStatusesData = []services.TaskStatusDetail{}
	}

	res := TasksStatusListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: taskCount,
		},
		Data: taskStatusesData,
	}
	c.JSON(res)
	return iris.StatusOK
}

// GetHistory 获取用户搜索历史
func (c *UserController) GetHistory() int {
	id := c.checkLogin()

	history := c.Service.GetSearchHistory(id)
	if history == nil {
		history = []string{}
	}

	c.JSON(struct {
		Data []string
	}{
		Data: history,
	})

	return iris.StatusOK
}

// DeleteHistory 清空用户搜索历史
func (c *UserController) DeleteHistory() int {
	id := c.checkLogin()
	c.Service.ClearSearchHistory(id)
	return iris.StatusOK
}

// FollowListRes 关注/粉丝列表数据
type FollowListRes struct {
	Pagination PaginationRes
	Data       []models.UserBaseInfo
}

// GetFollowerBy 获取用户粉丝列表
func (c *UserController) GetFollowerBy(id string) int {
	var userID primitive.ObjectID
	if id == "me" {
		userID = c.checkLogin()
	} else {
		var err error
		userID, err = primitive.ObjectIDFromHex(id)
		utils.AssertErr(err, "invalid_id", 400)
	}
	page, size := c.getPaginationData()

	followers, total := c.Service.GetFollower(userID, page, size)
	if followers == nil {
		followers = []models.UserBaseInfo{}
	}

	c.JSON(FollowListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: total,
		},
		Data: followers,
	})
	return iris.StatusOK
}

// GetFollowingBy 获取用户关注者列表
func (c *UserController) GetFollowingBy(id string) int {
	var userID primitive.ObjectID
	if id == "me" {
		userID = c.checkLogin()
	} else {
		var err error
		userID, err = primitive.ObjectIDFromHex(id)
		utils.AssertErr(err, "invalid_id", 400)
	}
	page, size := c.getPaginationData()

	followings, total := c.Service.GetFollowing(userID, page, size)
	if followings == nil {
		followings = []models.UserBaseInfo{}
	}

	c.JSON(FollowListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: total,
		},
		Data: followings,
	})

	return iris.StatusOK
}

// PostFollowingBy 添加关注
func (c *UserController) PostFollowingBy(id string) int {
	userID := c.checkLogin()
	followingID, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)
	c.Service.FollowUser(userID, followingID)
	return iris.StatusOK
}

// DeleteFollowingBy 取消关注
func (c *UserController) DeleteFollowingBy(id string) int {
	userID := c.checkLogin()
	followingID, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)
	c.Service.UnFollowUser(userID, followingID)
	return iris.StatusOK
}

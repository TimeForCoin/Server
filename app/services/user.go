package services

import (
	"time"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
)

// UserService 用户逻辑
type UserService interface {
	GetLoginURL() (url, state string)
	GetUser(id string) models.UserSchema
	UserAttend(id string)
	SetUserInfo(id string, info models.UserInfoSchema)
	LoginByCode(code string) (id string, new bool)
}

// NewUserService 初始化
func NewUserService() UserService {
	return &userService{
		model: models.GetModel().User,
		oAuth: libs.GetOAuth(),
	}
}

type userService struct {
	model *models.UserModel
	oAuth *libs.OAuthService
}

func (s *userService) GetLoginURL() (url, state string) {
	options := violet.AuthOption{
		Scopes:    violet.ScopeTypes{violet.ScopeInfo, violet.ScopeEmail},
		QuickMode: true,
	}
	url, state, err := s.oAuth.Api.GetLoginURL(s.oAuth.Callback, options)
	libs.Assert(err == nil, "Internal Server Error", iris.StatusInternalServerError)
	return url, state
}

func (s *userService) LoginByCode(code string) (id string, new bool) {
	res, err := s.oAuth.Api.GetToken(code)
	// TODO 检测是否绑定微信
	if err != nil {
		return "", false
	}
	// 账号已存在，直接返回 ID
	if u, err := s.model.GetUserByViolet(res.UserID); err == nil {
		return u.ID.Hex(), false
	}
	// 账号不存在，创建新账号
	if id, err = s.model.AddUserByViolet(res.UserID); err != nil {
		return "", false
	}
	// 获取用户信息
	info, err := s.oAuth.Api.GetUserInfo(res.Token)
	if err == nil {
		var birthday int64
		if t, err := time.Parse("2006-01-02T00:00:00.000Z", info.Birthday); err == nil {
			birthday = t.Unix()
		}
		gender := models.GenderMan
		if info.Gender == 1 {
			gender = models.GenderWoman
		} else if info.Gender == 2 {
			gender = models.GenderOther
		}
		_ = s.model.SetUserInfoByID(id, models.UserInfoSchema{
			Email:    info.Email,
			Phone:    info.Phone,
			Nickname: info.Nickname,
			Avatar:   info.Avatar,
			Bio:      info.Bio,
			Birthday: birthday,
			Gender:   gender,
			Location: info.Location,
		})
	}
	return "", true
}

// GetUser 获取用户数据
func (s *userService) GetUser(id string) models.UserSchema {
	user, err := s.model.GetUserByID(id)
	libs.Assert(err == nil, "faked_users", 403)
	return user
}

// UserAttend 用户签到
func (s *userService) UserAttend(id string) {
	user, err := s.model.GetUserByID(id)
	libs.Assert(err == nil, "invalid_session", 401)
	lastAttend := time.Unix(user.Data.AttendanceDate, 0)
	nowDate := time.Now()
	if lastAttend.Add(time.Hour*24).Before(nowDate) && lastAttend.YearDay() == nowDate.YearDay() {
		libs.Assert(false, "already_attend", 403)
	}
	err = s.model.SetUserAttend(id)
	libs.Assert(err == nil, err.Error(), iris.StatusInternalServerError)
}

func (s *userService) SetUserInfo(id string, info models.UserInfoSchema) {
	libs.Assert(s.model.SetUserInfoByID(id, info) == nil, "invalid_session", 401)
}

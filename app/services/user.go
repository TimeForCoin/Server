package services

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/xmatrixstudio/violet.sdk.go.v3"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
)

// UserService 用户逻辑
type UserService interface {
	GetLoginURL() (url, state string)
	GetUser(id primitive.ObjectID) models.UserSchema
	GetUserBaseInfo(id primitive.ObjectID) models.UserBaseInfo
	UserAttend(id primitive.ObjectID)
	SetUserInfo(id primitive.ObjectID, info models.UserInfoSchema)
	LoginByViolet(code string) (id string, new bool)
	LoginByWechat(code string) (id string, new bool)
	SetUserType(admin primitive.ObjectID, id primitive.ObjectID, userType models.UserType)
	// 认证相关
	CancelCertification(id primitive.ObjectID)
	UpdateCertification(id primitive.ObjectID, operate, data string)
	CheckCertification(id primitive.ObjectID, code string) string
	SendCertificationEmail(id primitive.ObjectID, email string)
	AddEmailCertification(identity models.UserIdentity, id primitive.ObjectID, data, email string)
	GetUserCollections(id primitive.ObjectID, page, size int64, sortRule string, taskType string,
		status string, reward string) (taskCount int64, taskCards []TaskDetail)
}

// NewUserService 初始化
func newUserService() UserService {
	return &userService{
		cache:     models.GetRedis().Cache,
		model:     models.GetModel().User,
		system:    models.GetModel().System,
		oAuth:     libs.GetOAuth(),
		taskModel: models.GetModel().Task,
		setModel:  models.GetModel().Set,
		fileModel: models.GetModel().File,
	}
}

type userService struct {
	cache     *models.CacheModel
	system    *models.SystemModel
	model     *models.UserModel
	oAuth     *libs.OAuthService
	taskModel *models.TaskModel
	setModel  *models.SetModel
	fileModel *models.FileModel
}

func (s *userService) GetUserBaseInfo(id primitive.ObjectID) models.UserBaseInfo {
	user, err := s.cache.GetUserBaseInfo(id)
	if err != nil {
		return models.UserBaseInfo{
			ID:       primitive.NewObjectID().Hex(),
			Avatar:   "",
			Nickname: "匿名用户",
			Gender:   models.GenderOther,
			Type:     models.UserTypeBan,
		}
	}
	return user
}

func (s *userService) GetLoginURL() (url, state string) {
	options := violet.AuthOption{
		Scopes:    violet.ScopeTypes{violet.ScopeInfo, violet.ScopeEmail},
		QuickMode: true,
	}
	url, state, err := s.oAuth.Api.GetLoginURL(s.oAuth.Callback, options)
	libs.Assert(err == nil, "Internal Service Error", iris.StatusInternalServerError)
	return url, state
}

func (s *userService) LoginByViolet(code string) (id string, new bool) {
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
	userID, err := s.model.AddUserByViolet(res.UserID)
	if err != nil {
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
		// 保存头像到云存储
		url, err := libs.GetCOS().SaveURLFile("avatar-" + userID.Hex() + ".png", info.Avatar)
		libs.AssertErr(err, "", 500)
		_ = s.model.SetUserInfoByID(userID, models.UserInfoSchema{
			Email:    info.Email,
			Phone:    info.Phone,
			Nickname: info.Nickname,
			Avatar:   url,
			Bio:      info.Bio,
			Birthday: birthday,
			Gender:   gender,
			Location: info.Location,
		})
	}
	return userID.Hex(), true
}

func (s *userService) LoginByWechat(code string) (id string, new bool) {
	openID, err := libs.GetWechat().GetOpenID(code)
	libs.AssertErr(err, "", 403)
	// 账号已存在，直接返回 ID
	if u, err := s.model.GetUserByWechat(openID); err == nil {
		return u.ID.Hex(), u.Info.Nickname == ""
	}
	// 账号不存在，新建账号
	userID, err := s.model.AddUserByWechat(openID)
	libs.AssertErr(err, "db_error", iris.StatusInternalServerError)
	return userID.Hex(), true
}

// GetUser 获取用户数据
func (s *userService) GetUser(id primitive.ObjectID) models.UserSchema {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "faked_users", 403)
	return user
}

// UserAttend 用户签到
func (s *userService) UserAttend(id primitive.ObjectID) {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "invalid_session", 401)
	lastAttend := time.Unix(user.Data.AttendanceDate, 0)
	nowDate := time.Now()
	if lastAttend.Add(time.Hour*24).After(nowDate) && lastAttend.YearDay() == nowDate.YearDay() {
		libs.Assert(false, "already_attend", 403)
	}
	libs.Assert(s.model.SetUserAttend(id) == nil, "unknown", iris.StatusInternalServerError)
}

// 设置用户信息
func (s *userService) SetUserInfo(id primitive.ObjectID, info models.UserInfoSchema) {
	libs.Assert(s.model.SetUserInfoByID(id, info) == nil, "invalid_session", 401)
	libs.Assert(models.GetRedis().Cache.WillUpdate(id, models.KindOfBaseInfo) == nil, "redis_error", iris.StatusInternalServerError)
}

// 设置用户类型
func (s *userService) SetUserType(admin primitive.ObjectID, id primitive.ObjectID, userType models.UserType) {
	adminInfo, err := models.GetRedis().Cache.GetUserBaseInfo(admin)
	libs.AssertErr(err, "invalid_session", 401)
	libs.Assert(adminInfo.Type == models.UserTypeAdmin ||
		adminInfo.Type == models.UserTypeRoot, "permission_deny", 403)
	err = s.model.SetUserType(id, userType)
	libs.Assert(err == nil, "faked_users", 403)
	libs.Assert(models.GetRedis().Cache.WillUpdate(id, models.KindOfBaseInfo) == nil, "redis_error", iris.StatusInternalServerError)
}

// 取消认证
func (s *userService) CancelCertification(id primitive.ObjectID) {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "invalid_session", 401)
	libs.Assert(user.Certification.Identity != models.IdentityNone, "faked_certification")
	user.Certification.Status = models.CertificationCancel
	err = s.model.SetUserCertification(id, user.Certification)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

// 更新认证[管理员]
func (s *userService) UpdateCertification(id primitive.ObjectID, operate, data string) {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "faked_user", 401)
	libs.Assert(user.Certification.Identity != models.IdentityNone, "faked_certification")
	if operate == "true" {
		if data != "" {
			user.Certification.Data = data
		}
		user.Certification.Date = time.Now().Unix()
		user.Certification.Status = models.CertificationTrue
		err = s.model.SetUserCertification(id, user.Certification)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
	} else if operate == "false" {
		user.Certification.Status = models.CertificationFalse
		user.Certification.Feedback = data
		err = s.model.SetUserCertification(id, user.Certification)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
	}
}

// AddCertification 添加认证
func (s *userService) AddEmailCertification(identity models.UserIdentity, id primitive.ObjectID, data, email string) {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "invalid_session", 401)
	libs.Assert(user.Certification.Identity == models.IdentityNone ||
		user.Certification.Status == models.CertificationCancel, "exist_certification", 403)
	libs.Assert(s.model.CheckCertificationEmail(email), "exist_email", 403)
	err = s.model.SetUserCertification(id, models.UserCertificationSchema{
		Identity: identity,
		Data:     data,
		Status:   models.CertificationCheckEmail,
		Date:     time.Now().Unix(),
		Email:    email,
	})
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	s.SendCertificationEmail(id, email)
}

// 发送认证邮件
func (s *userService) SendCertificationEmail(id primitive.ObjectID, email string) {
	if email == "" {
		user, err := s.model.GetUserByID(id)
		libs.AssertErr(err, "invalid_session", 401)
		libs.Assert(user.Certification.Identity != models.IdentityNone &&
			user.Certification.Status == models.CertificationCheckEmail, "faked_certification", 403)
		libs.Assert(user.Certification.Email != "", "faked_email", 403)
		email = user.Certification.Email
	}
	exist, _ := models.GetRedis().Cache.CheckCertification(id, "", "", false)
	libs.Assert(!exist, "limit_email", 403)
	token := libs.GetRandomString(64)
	code := libs.GetHash(token + "&" + email)
	err := libs.GetEmail().SendAuthEmail(id, email, code)
	libs.AssertErr(err, "error_email", iris.StatusInternalServerError)
	err = models.GetRedis().Cache.SetCertification(id, token)
	libs.AssertErr(err, "error_redis", iris.StatusInternalServerError)
}

func (s *userService) CheckCertification(id primitive.ObjectID, code string) string {
	user, err := s.model.GetUserByID(id)
	if err != nil {
		return "无效的认证链接"
	}
	// 是否存在认证
	libs.Assert(user.Certification.Identity != models.IdentityNone, "error_code", iris.StatusInternalServerError)
	// 检查认证邮箱和 code 的正确性
	exist, right := models.GetRedis().Cache.CheckCertification(id, user.Certification.Email, code, true)
	if !exist || !right || user.Certification.Status != models.CertificationCheckEmail {
		return "无效的认证链接"
	}
	// 是否自动通过认证
	emailParts := strings.Split(user.Certification.Email, "@")
	if len(emailParts) != 2 {
		return "无效的认证邮箱"
	}
	content := s.system.ExistAutoEmail(emailParts[1])
	if content == "" {
		user.Certification.Status = models.CertificationWaitEmail
		err = s.model.SetUserCertification(id, user.Certification)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		return "认证邮箱成功，待审核通过"
	}
	user.Certification.Status = models.CertificationTrue
	user.Certification.Data = content
	err = s.model.SetUserCertification(id, user.Certification)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	return "认证已通过"
}

func (s *userService) GetUserCollections(id primitive.ObjectID, page, size int64, sortRule string, taskType string,
	status string, reward string) (taskCount int64, taskCards []TaskDetail) {
	var taskTypes []models.TaskType
	var statuses []models.TaskStatus
	var rewards []models.RewardType
	var keywords []string
	split := strings.Split(taskType, ",")
	for _, str := range split {
		if str == "all" {
			taskTypes = []models.TaskType{models.TaskTypeRunning, models.TaskTypeQuestionnaire, models.TaskTypeInfo}
			break
		}
		taskTypes = append(taskTypes, models.TaskType(str))
	}
	split = strings.Split(status, ",")
	for _, str := range split {
		if str == "all" {
			statuses = []models.TaskStatus{models.TaskStatusClose, models.TaskStatusFinish, models.TaskStatusWait}
			break
		}
		statuses = append(statuses, models.TaskStatus(str))
	}
	split = strings.Split(reward, ",")
	for _, str := range split {
		if str == "all" {
			rewards = []models.RewardType{models.RewardMoney, models.RewardObject, models.RewardRMB}
			break
		}
		rewards = append(rewards, models.RewardType(str))
	}

	if sortRule == "new" {
		sortRule = "publish_date"
	}
	collectionTasks := s.setModel.GetSets(id, "collect_task_id")
	tasks, taskCount, err := s.taskModel.GetTasks(sortRule, collectionTasks.CollectTaskID, taskTypes, statuses, rewards, keywords, "", (page-1)*size, size)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	for i, t := range tasks {
		var task TaskDetail
		task.TaskSchema = &tasks[i]

		user, err := s.cache.GetUserBaseInfo(t.Publisher)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		task.Publisher = user

		images, err := s.fileModel.GetFileByContent(t.ID, models.FileImage)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		task.Images = []ImagesData{}
		task.Attachment = []models.FileSchema{}
		for _, i := range images {
			task.Images = append(task.Images, ImagesData{
				ID:  i.ID.Hex(),
				URL: i.URL,
			})
		}
		task.Liked = s.cache.IsLikeTask(id, t.ID)
		task.Collected = s.cache.IsCollectTask(id, t.ID)

		taskCards = append(taskCards, task)
	}

	return
}

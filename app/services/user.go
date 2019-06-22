package services

import (
	"math/rand"
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
	GetUser(id primitive.ObjectID, isMe bool) UserDetail
	GetUserBaseInfo(id primitive.ObjectID) models.UserBaseInfo
	UserAttend(id primitive.ObjectID)
	SetUserInfo(id primitive.ObjectID, info models.UserInfoSchema)
	LoginByViolet(code string) (id string, new bool)
	LoginByWechat(code string) (id string, new bool)
	SetUserType(admin primitive.ObjectID, id primitive.ObjectID, userType models.UserType)
	SearchUser(key string, page, size int64) []UserDetail
	GetUserCollections(id primitive.ObjectID, page, size int64, sortRule string, taskType string,
		status string, reward string) (taskCount int64, taskCards []TaskDetail)
	GetUserParticipate(id primitive.ObjectID, page, size int64, status string) (taskStatusCount int64, taskStatusDetailList []TaskStatusDetail)
	// 搜索相关
	GetSearchHistory(id primitive.ObjectID) []string
	ClearSearchHistory(id primitive.ObjectID)
	// 认证相关
	CancelCertification(id primitive.ObjectID)
	UpdateCertification(id primitive.ObjectID, operate, data string)
	CheckCertification(id primitive.ObjectID, code string) string
	SendCertificationEmail(id primitive.ObjectID, email string)
	AddEmailCertification(identity models.UserIdentity, id primitive.ObjectID, data, email string)
	AddMaterialCertification(identity models.UserIdentity, id primitive.ObjectID, data string, attachment []primitive.ObjectID)
	GetCertificationList(userID primitive.ObjectID, types []models.CertificationStatus, page, size int64) (users []UserDetail)
	GetAutoCertification(userID primitive.ObjectID, page, size int64) (keys []models.SystemSchemas)
	AddAutoCertification(userID primitive.ObjectID, key, data string)
	RemoveAutoCertification(userID primitive.ObjectID, key string)
	// 关注相关
	GetFollowing(id primitive.ObjectID, page, size int64) ([]models.UserBaseInfo, int64)
	GetFollower(id primitive.ObjectID, page, size int64) ([]models.UserBaseInfo, int64)
	FollowUser(userID, followID primitive.ObjectID)
	UnFollowUser(userID, followID primitive.ObjectID)
	IsFollower(userID, followID primitive.ObjectID) bool
	IsFollowing(userID, followID primitive.ObjectID) bool
}

// NewUserService 初始化
func newUserService() UserService {
	return &userService{
		cache:           models.GetRedis().Cache,
		model:           models.GetModel().User,
		system:          models.GetModel().System,
		oAuth:           libs.GetOAuth(),
		taskModel:       models.GetModel().Task,
		setModel:        models.GetModel().Set,
		fileModel:       models.GetModel().File,
		taskStatusModel: models.GetModel().TaskStatus,
	}
}

type userService struct {
	cache           *models.CacheModel
	system          *models.SystemModel
	model           *models.UserModel
	oAuth           *libs.OAuthService
	taskModel       *models.TaskModel
	setModel        *models.SetModel
	fileModel       *models.FileModel
	taskStatusModel *models.TaskStatusModel
}

// UserDetail 用户详细信息
type UserDetail struct {
	UserID        primitive.ObjectID `json:"-"`
	ID            string             `json:"id"`
	VioletName    string             `json:"violet_name,omitempty"`
	WechatName    string             `json:"wechat_name,omitempty"`
	RegisterTime  int64
	Info          models.UserInfoSchema
	Data          *UserDataRes
	Certification *UserCertification
}

// UserDataRes 用户数据返回值
type UserDataRes struct {
	*models.UserDataSchema
	// 额外项
	Attendance   bool  // 是否签到
	CollectCount int64 // 收藏任务数
	Follower     bool  // 是否为自己的粉丝
	Following    bool  // 是否已关注
	// 排除项
	AttendanceDate omit `json:"attendance_date,omitempty"`
	CollectTasks   omit `json:"collect_tasks,omitempty"`
	SearchHistory  omit `json:"search_history,omitempty"`
}

// UserCertification 用户认证信息
type UserCertification struct {
	Type     models.UserIdentity
	Status   models.CertificationStatus
	Email    string
	Data     string
	Date     int64
	Material []models.FileSchema `json:"material,omitempty"`
}

// TaskStatus 任务参与情况
type TaskStatus struct {
	*models.TaskStatusSchema
	// 额外项
	Player models.UserBaseInfo
	//排除项
	ID omit `json:"id,omitempty"` // 点赞用户ID
}

// TaskStatusDetail 任务参与详情
type TaskStatusDetail struct {
	Status TaskStatus
	Task   TaskDetail
}

// GetUserBaseInfo 获取用户基本信息
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

// GetLoginURL 获取登陆链接
func (s *userService) GetLoginURL() (url, state string) {
	options := violet.AuthOption{
		Scopes:    violet.ScopeTypes{violet.ScopeInfo, violet.ScopeEmail},
		QuickMode: true,
	}
	url, state, err := s.oAuth.API.GetLoginURL(s.oAuth.Callback, options)
	libs.Assert(err == nil, "Internal Service Error", iris.StatusInternalServerError)
	return url, state
}

// LoginByViolet 使用 Violet 授权登陆
func (s *userService) LoginByViolet(code string) (id string, new bool) {
	res, err := s.oAuth.API.GetToken(code)
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
	info, err := s.oAuth.API.GetUserInfo(res.Token)
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
		url, err := libs.GetCOS().SaveURLFile("avatar-"+userID.Hex()+".png", info.Avatar)
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

// LoginByWechat 使用微信登陆
func (s *userService) LoginByWechat(code string) (id string, new bool) {
	openID, err := libs.GetWeChat().GetOpenID(code)
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
func (s *userService) GetUser(id primitive.ObjectID, isMe bool) UserDetail {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "faked_users", 403)
	return s.makeUserRes(user, isMe)
}

// SearchUser 搜索用户
func (s *userService) SearchUser(key string, page, size int64) (res []UserDetail) {
	users := s.model.GetUsers(key, page, size)
	for i := range users {
		res = append(res, s.makeUserRes(users[i], false))
	}
	return res
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
	err = s.model.UpdateUserDataCount(id, models.UserDataCount{
		Value: rand.Int63n(20),
	})
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	err = s.model.SetUserAttend(id)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

// SetUserInfo 设置用户信息
func (s *userService) SetUserInfo(id primitive.ObjectID, info models.UserInfoSchema) {
	libs.Assert(s.model.SetUserInfoByID(id, info) == nil, "invalid_session", 401)
	libs.Assert(models.GetRedis().Cache.WillUpdate(id, models.KindOfBaseInfo) == nil, "redis_error", iris.StatusInternalServerError)
}

// SetUserType 设置用户类型
func (s *userService) SetUserType(admin primitive.ObjectID, id primitive.ObjectID, userType models.UserType) {
	adminInfo, err := models.GetRedis().Cache.GetUserBaseInfo(admin)
	libs.AssertErr(err, "invalid_session", 401)
	libs.Assert(adminInfo.Type == models.UserTypeAdmin ||
		adminInfo.Type == models.UserTypeRoot, "permission_deny", 403)
	err = s.model.SetUserType(id, userType)
	libs.Assert(err == nil, "faked_users", 403)
	libs.Assert(models.GetRedis().Cache.WillUpdate(id, models.KindOfBaseInfo) == nil, "redis_error", iris.StatusInternalServerError)
}

// CancelCertification 取消认证
func (s *userService) CancelCertification(id primitive.ObjectID) {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "invalid_session", 401)
	libs.Assert(user.Certification.Identity != models.IdentityNone, "faked_certification")
	user.Certification.Status = models.CertificationCancel
	err = s.model.SetUserCertification(id, user.Certification)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

// UpdateCertification 更新认证[管理员]
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

// AddEmailCertification 添加认证
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

// AddMaterialCertification 添加材料认证
func (s *userService) AddMaterialCertification(identity models.UserIdentity, id primitive.ObjectID, data string, attachment []primitive.ObjectID) {
	user, err := s.model.GetUserByID(id)
	libs.AssertErr(err, "invalid_session", 401)
	libs.Assert(user.Certification.Identity == models.IdentityNone ||
		user.Certification.Status == models.CertificationCancel, "exist_certification", 403)
	GetServiceManger().File.BindFilesToUser(id, attachment)
	err = s.model.SetUserCertification(id, models.UserCertificationSchema{
		Identity: identity,
		Data:     data,
		Status:   models.CertificationWait,
		Date:     time.Now().Unix(),
		Material: attachment,
	})
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

// SendCertificationEmail 发送认证邮件
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

// CheckCertification 检查用户认证
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

// GetCertificationList 获取待审核认证列表
func (s *userService) GetCertificationList(userID primitive.ObjectID, types []models.CertificationStatus, page, size int64) (users []UserDetail) {
	admin := s.GetUserBaseInfo(userID)
	libs.Assert(admin.Type == models.UserTypeAdmin || admin.Type == models.UserTypeRoot, "permission_deny", 403)

	usersData, err := s.model.GetCertification(types, page, size)
	libs.AssertErr(err, "", 500)
	for i := range usersData {
		users = append(users, s.makeUserRes(usersData[i], true))
	}
	return users
}

// GetAutoCertification 获取自动认证后缀
func (s *userService) GetAutoCertification(userID primitive.ObjectID, page, size int64) (keys []models.SystemSchemas) {
	admin := s.GetUserBaseInfo(userID)
	libs.Assert(admin.Type == models.UserTypeAdmin || admin.Type == models.UserTypeRoot, "permission_deny", 403)
	keys, err := s.system.GetAutoEmail(page, size)
	libs.AssertErr(err, "", 500)
	for i := range keys {
		keys[i].Key = strings.Replace(keys[i].Key, "email-", "", 1)
	}
	return keys
}

// AddAutoCertification 添加自动认证后缀
func (s *userService) AddAutoCertification(userID primitive.ObjectID, key, data string) {
	admin := s.GetUserBaseInfo(userID)
	libs.Assert(admin.Type == models.UserTypeAdmin || admin.Type == models.UserTypeRoot, "permission_deny", 403)
	err := s.system.AddAutoEmail(key, data)
	libs.AssertErr(err, "", 500)
}

// RemoveAutoCertification 移除自动认证后缀
func (s *userService) RemoveAutoCertification(userID primitive.ObjectID, key string) {
	admin := s.GetUserBaseInfo(userID)
	libs.Assert(admin.Type == models.UserTypeAdmin || admin.Type == models.UserTypeRoot, "permission_deny", 403)
	err := s.system.RemoveAutoEmail(key)
	libs.AssertErr(err, "faked_email", 403)
}

// GetUserCollections 获取用户收藏
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

	collectionTasks := s.setModel.GetSets(id, models.SetOfCollectTask)
	if len(collectionTasks.CollectTaskID) > 0 {
		tasks, taskCount, err := s.taskModel.GetTasks(sortRule, collectionTasks.CollectTaskID, taskTypes, statuses, rewards, keywords, "", (page-1)*size, size)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		for _, t := range tasks {
			taskCards = append(taskCards, GetServiceManger().Task.makeTaskDetail(t, id.Hex()))
		}
		return taskCount, taskCards
	}
	return 0, []TaskDetail{}
}

// GetSearchHistory 获取用户搜索历史
func (s *userService) GetSearchHistory(id primitive.ObjectID) []string {
	res, err := s.model.GetSearchHistory(id)
	libs.AssertErr(err, "", 500)
	return res
}

// ClearSearchHistory 清空用户搜索历史
func (s *userService) ClearSearchHistory(id primitive.ObjectID) {
	err := s.model.ClearSearchHistory(id)
	libs.AssertErr(err, "", 500)
}

// GetFollowing 获取用户关注列表
func (s *userService) GetFollowing(id primitive.ObjectID, page, size int64) ([]models.UserBaseInfo, int64) {
	set := s.setModel.GetSets(id, models.SetOfFollowingUser)
	followingIDs := set.FollowingUserID
	lenOfIDs := int64(len(followingIDs))
	beginIndex := (page - 1) * size
	endIndex := page * size
	if lenOfIDs < beginIndex {
		return []models.UserBaseInfo{}, lenOfIDs
	}
	if lenOfIDs > endIndex {
		followingIDs = followingIDs[beginIndex:endIndex]
	} else {
		followingIDs = followingIDs[beginIndex:]
	}
	var res []models.UserBaseInfo
	for _, id := range followingIDs {
		user, _ := s.cache.GetUserBaseInfo(id)
		res = append(res, user)
	}
	return res, lenOfIDs
}

// GetFollower 获取用户粉丝列表
func (s *userService) GetFollower(id primitive.ObjectID, page, size int64) ([]models.UserBaseInfo, int64) {
	set := s.setModel.GetSets(id, models.SetOfFollowerUser)
	followerIDs := set.FollowerUserID
	lenOfIDs := int64(len(followerIDs))
	beginIndex := (page - 1) * size
	endIndex := page * size
	if lenOfIDs < beginIndex {
		return []models.UserBaseInfo{}, lenOfIDs
	}
	if lenOfIDs > endIndex {
		followerIDs = followerIDs[beginIndex:endIndex]
	} else {
		followerIDs = followerIDs[beginIndex:]
	}
	var res []models.UserBaseInfo
	for _, id := range followerIDs {
		user, _ := s.cache.GetUserBaseInfo(id)
		res = append(res, user)
	}
	return res, lenOfIDs
}

// FollowUser 关注用户
func (s *userService) FollowUser(userID, followID primitive.ObjectID) {
	_, err := s.model.GetUserByID(followID)
	libs.AssertErr(err, "faked_user", 403)

	err = s.setModel.AddToSet(userID, followID, models.SetOfFollowingUser)
	libs.AssertErr(err, "exist_relation", 403)

	err = s.setModel.AddToSet(followID, userID, models.SetOfFollowerUser)
	libs.AssertErr(err, "", 500)

	err = s.model.UpdateUserDataCount(followID, models.UserDataCount{
		FollowerCount: 1,
	})
	libs.AssertErr(err, "", 500)
	err = s.model.UpdateUserDataCount(userID, models.UserDataCount{
		FollowingCount: 1,
	})
	libs.AssertErr(err, "", 500)

	err = s.cache.WillUpdate(userID, models.KindOfFollowing)
	libs.AssertErr(err, "", 500)

	err = s.cache.WillUpdate(followID, models.KindOfFollower)
	libs.AssertErr(err, "", 500)
}

func (s *userService) GetUserParticipate(id primitive.ObjectID, page, size int64, status string) (taskStatusCount int64, taskStatusDetailList []TaskStatusDetail) {
	var statuses []models.PlayerStatus
	split := strings.Split(status, ",")
	for _, str := range split {
		if str == "all" {
			statuses = []models.PlayerStatus{models.PlayerWait, models.PlayerRefuse, models.PlayerClose, models.PlayerRunning, models.PlayerFinish, models.PlayerGiveUp, models.PlayerFailure}
			break
		}
		statuses = append(statuses, models.PlayerStatus(str))
	}
	taskStatusList, taskStatusCount, err := s.taskStatusModel.GetTaskStatusListByUserID(id, statuses, (page-1)*size, size)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	userPlayer := s.GetUserBaseInfo(id)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	for i := range taskStatusList {
		var taskStatusDetail TaskStatusDetail
		taskStatusDetail.Status = TaskStatus{
			TaskStatusSchema: &taskStatusList[i],
			Player: userPlayer,
		}
		task, err := s.taskModel.GetTaskByID(taskStatusList[i].ID)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		taskStatusDetail.Task = GetServiceManger().Task.makeTaskDetail(task, id.Hex())
		taskStatusDetailList = append(taskStatusDetailList, taskStatusDetail)
	}
	return
}

// UnFollowUser 取消关注用户
func (s *userService) UnFollowUser(userID, followID primitive.ObjectID) {
	err := s.setModel.RemoveFromSet(userID, followID, models.SetOfFollowingUser)
	libs.AssertErr(err, "faked_relation", 403)

	err = s.setModel.RemoveFromSet(followID, userID, models.SetOfFollowerUser)
	libs.AssertErr(err, "", 500)

	err = s.model.UpdateUserDataCount(followID, models.UserDataCount{
		FollowerCount: -1,
	})
	libs.AssertErr(err, "", 500)
	err = s.model.UpdateUserDataCount(userID, models.UserDataCount{
		FollowingCount: -1,
	})
	libs.AssertErr(err, "", 500)

	err = s.cache.WillUpdate(userID, models.KindOfFollowing)
	libs.AssertErr(err, "", 500)

	err = s.cache.WillUpdate(followID, models.KindOfFollower)
	libs.AssertErr(err, "", 500)
}

// IsFollower 是否是粉丝
func (s *userService) IsFollower(userID, followID primitive.ObjectID) bool {
	return s.cache.IsFollowerUser(userID, followID)
}

// IsFollowing 是否已关注
func (s *userService) IsFollowing(userID, followID primitive.ObjectID) bool {
	return s.cache.IsFollowingUser(userID, followID)
}

// makeUserRes 整合用户数据
func (s *userService) makeUserRes(user models.UserSchema, all bool) UserDetail {
	res := UserDetail{
		UserID:       user.ID,
		ID:           user.ID.Hex(),
		VioletName:   user.VioletName,
		WechatName:   user.WechatName,
		RegisterTime: user.RegisterTime,
		Info:         user.Info,
		Data: &UserDataRes{
			UserDataSchema: &user.Data,
			CollectCount:   int64(len(user.Data.CollectTasks)),
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
	if !all {
		res.Certification.Email = ""
		if res.Certification.Status != models.CertificationTrue {
			res.Certification = &UserCertification{
				Type: models.IdentityNone,
			}
		}
	} else {
		res.Certification.Material = []models.FileSchema{}
		for _, id := range user.Certification.Material {
			file, err := s.fileModel.GetFile(id)
			libs.AssertErr(err, "", 500)
			res.Certification.Material = append(res.Certification.Material, file)
		}
	}
	return res
}

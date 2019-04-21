package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserModel User 数据库
type UserModel struct {
	Collection *mongo.Collection
}

// UserGender 用户性别
type UserGender string

// UserType 用户类型
type UserType string

// UserIdentity 用户认证身份
type UserIdentity string

// UserLoginType 用户登陆类型
type UserLoginType string

// CertificationStatus 用户认证状态
type CertificationStatus string

// UserGender 用户性别
const (
	GenderMan   UserGender = "man"   // 男
	GenderWoman UserGender = "woman" // 女
	GenderOther UserGender = "other" // 其他
)

// UserType 用户类型
const (
	UserTypeBan    UserType = "ban"    // 禁封用户
	UserTypeWeChat UserType = "wechat" // 微信临时用户
	UserTypeNormal UserType = "normal" // 正式用户
	UserTypeAdmin  UserType = "admin"  // 管理员
	UserTypeRoot   UserType = "root"   // 超级管理员
)

// UserIdentity 用户认证身份
const (
	IdentityStudent UserIdentity = "student" // 目前仅支持学生认证
)

// UserLoginType 用户登陆类型
const (
	LoginByWeb    UserLoginType = "web"    // 网页登陆
	LoginByWeChat UserLoginType = "wechat" // 微信登陆
)

// CertificationStatus 用户认证状态
const (
	CertificationTrue       CertificationStatus = "true"        // 认证通过
	CertificationFalse      CertificationStatus = "false"       // 认证失败
	CertificationWait       CertificationStatus = "wait"        // 认证材料已提交，待审核
	CertificationCheckEmail CertificationStatus = "check_email" // 认证邮件已发送，待确认
	CertificationWaitEmail  CertificationStatus = "wait_email"  // 认证邮件已发确认，待审核
	CertificationCancel     CertificationStatus = "cancel"      // 认证已被用户取消
)

// UserInfoSchema 用户信息结构
type UserInfoSchema struct {
	Email    string     // 联系邮箱
	Phone    string     // 联系手机
	Avatar   string     // 头像
	Nickname string     // 用户昵称
	Bio      string     // 个人简介
	Gender   UserGender // 性别
	Location string     // 具体位置
	BirthDay int64      // 生日
}

// UserDataSchema 用户数据结构
type UserDataSchema struct {
	Money          int64                // 当前持有闲币
	Value          int64                // 用户积分
	Credit         int64                // 个人信誉
	Type           UserType             // 用户类型
	AttendanceDate int64                `bson:"attendance_date"` // 签到时间戳
	CollectTasks   []primitive.ObjectID `bson:"collect_tasks"`   // 收藏的任务
	SearchHistory  []string             `bson:"searchHistory"`   // 搜索历史(仅保留最近的 20 条)
	// 冗余数据
	LastLoginDate   int64         `bson:"last_login_date"`   // 上次登陆时间(日志有)
	LastLoginType   UserLoginType `bson:"last_login_type"`   // 上次登陆类型（PC/微信小程序）
	PublishCount    int64         `bson:"publish_count"`     // 发布任务数
	PublishRunCount int64         `bson:"publish_run_count"` // 发布并进行中任务数
	ReceiveCount    int64         `bson:"receive_count"`     // 领取任务数
	ReceiveRunCount int64         `bson:"receive_run_count"` // 领取并进行中任务数
	FollowingCount  int64         `bson:"following_count"`   // 关注人数量
	FollowerCount   int64         `bson:"follower_count"`    // 粉丝数量
}

// UserCertificationSchema 用户认证信息
type UserCertificationSchema struct {
	Identity   UserIdentity         `bson:"identity"`    // 认证身份类型
	Data       string               `bson:"data"`        // 认证内容
	Status     CertificationStatus  `bson:"status"`      // 认证状态
	Date       int64                `bson:"date"`        // 认证时间
	Material   []primitive.ObjectID `bson:"material"`    // 人工认证材料
	Feedback   string               `bson:"feedback"`    // 审核不通过后的反馈
	Email      string               `bson:"email"`       // 邮箱认证
	EmailToken string               `bson:"email_token"` // 邮箱认证Token
}

// UserSchema User 基本数据结构
type UserSchema struct {
	ID            primitive.ObjectID      `bson:"_id,omitempty"` // 用户ID [索引]
	OpenID        string                  `bson:"open_id"`       // 微信OpenID
	WeChatName    string                  `bson:"wechat_name"`   // 微信名
	VioletID      string                  `bson:"violet_id"`     // VioletID
	Name          string                  `bson:"name"`          // 用户名， 唯一
	RegisterTime  int64                   `bson:"register_time"` // 用户注册时间
	Info          UserInfoSchema          `bson:"info"`          // 用户个性信息
	Data          UserDataSchema          `bson:"data"`          // 用户数据
	Certification UserCertificationSchema `bson:"certification"` // 用户认证信息
}

// AddUser 增加用户
func (model *UserModel) AddUser(name string) error {
	ctx, over := GetCtx()
	defer over()
	// 返回ID
	_, err := model.Collection.InsertOne(ctx, &UserSchema{Name: name})
	return err
}

// FindUser 查找用户
func (model *UserModel) FindUser(name string) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = model.Collection.FindOne(ctx, bson.M{"name": name}).Decode(&user)
	return
}

// UpdateUser 更新用户
func (model *UserModel) UpdateUser(name string, newName string) error {
	ctx, over := GetCtx()
	defer over()
	// 返回更新的数量
	res, err := model.Collection.UpdateOne(ctx, bson.M{"name": name}, bson.M{"$set": bson.M{"name": newName}})
	if err != nil {
		return nil
	}
	if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// RemoveUser 删除用户
func (model *UserModel) RemoveUser(name string) error {
	ctx, over := GetCtx()
	defer over()
	res, err := model.Collection.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return ErrNotExist
	}
	return nil
}

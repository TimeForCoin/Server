package models

import (
	"errors"
	"reflect"
	"time"

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
	UserTypeNormal UserType = "normal" // 正式用户
	UserTypeAdmin  UserType = "admin"  // 管理员
	UserTypeRoot   UserType = "root"   // 超级管理员
)

// UserIdentity 用户认证身份
const (
	IdentityNone    UserIdentity = "none"    // 未认证
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
	Email    string     `bson:"email"`    // 联系邮箱
	Phone    string     `bson:"phone"`    // 联系手机
	Avatar   string     `bson:"avatar"`   // 头像
	School   string     `bson:"school"`   // 学校
	Nickname string     `bson:"nickname"` // 用户昵称
	Bio      string     `bson:"bio"`      // 个人简介
	Gender   UserGender `bson:"gender"`   // 性别
	Location string     `bson:"location"` // 具体位置
	Birthday int64      `bson:"birthday"` // 生日
}

// UserDataSchema 用户数据结构
type UserDataSchema struct {
	Money  int64    // 当前持有闲币
	Value  int64    // 用户积分
	Credit int64    // 个人信誉
	Type   UserType // 用户类型

	AttendanceDate int64                `bson:"attendance_date"` // 签到时间戳
	CollectTasks   []primitive.ObjectID `bson:"collect_tasks"`   // 收藏的任务
	SearchHistory  []string             `bson:"searchHistory"`   // 搜索历史(仅保留最近的 20 条)
	// 冗余数据
	PublishCount    int64 `bson:"publish_count"`     // 发布任务数
	PublishRunCount int64 `bson:"publish_run_count"` // 发布并进行中任务数
	ReceiveCount    int64 `bson:"receive_count"`     // 领取任务数
	ReceiveRunCount int64 `bson:"receive_run_count"` // 领取并进行中任务数
	FollowingCount  int64 `bson:"following_count"`   // 关注人数量
	FollowerCount   int64 `bson:"follower_count"`    // 粉丝数量
}

// UserCertificationSchema 用户认证信息
type UserCertificationSchema struct {
	Identity UserIdentity         `bson:"identity"` // 认证身份类型
	Data     string               `bson:"data"`     // 认证内容
	Status   CertificationStatus  `bson:"status"`   // 认证状态
	Date     int64                `bson:"date"`     // 认证时间
	Material []primitive.ObjectID `bson:"material"` // 人工认证材料
	Feedback string               `bson:"feedback"` // 审核不通过后的反馈
	Email    string               `bson:"email"`    // 邮箱认证
}

// UserSchema User 基本数据结构
type UserSchema struct {
	ID            primitive.ObjectID      `bson:"_id,omitempty"` // 用户ID [索引]
	WechatID      string                  `bson:"wechat_id"`     // 微信OpenID
	WechatName    string                  `bson:"wechat_name"`   // 微信名
	VioletID      string                  `bson:"violet_id"`     // VioletID
	VioletName    string                  `bson:"violet_name"`   // Violet 用户名
	RegisterTime  int64                   `bson:"register_time"` // 用户注册时间
	Info          UserInfoSchema          `bson:"info"`          // 用户个性信息
	Data          UserDataSchema          `bson:"data"`          // 用户数据
	Certification UserCertificationSchema `bson:"certification"` // 用户认证信息
}

// AddUserByViolet 通过 Violet 增加用户
func (model *UserModel) AddUserByViolet(id string) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()
	userID := primitive.NewObjectID()
	_, err := model.Collection.InsertOne(ctx, &UserSchema{
		ID:           userID,
		VioletID:     id,
		RegisterTime: time.Now().Unix(),
		Data: UserDataSchema{
			Type: UserTypeNormal,
		},
		Certification: UserCertificationSchema{
			Identity: IdentityNone,
		},
	})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return userID, nil
}

func (model *UserModel) AddUserByWechat(openid string) (primitive.ObjectID, error) {
	ctx, finish := GetCtx()
	defer finish()
	userID := primitive.NewObjectID()
	_, err := model.Collection.InsertOne(ctx, &UserSchema{
		ID:           userID,
		WechatID:     openid,
		RegisterTime: time.Now().Unix(),
		Data: UserDataSchema{
			Type: UserTypeNormal,
		},
		Certification: UserCertificationSchema{
			Identity: IdentityNone,
		},
	})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return userID, nil
}

// GetUserByID 通过 ID 查找用户
func (model *UserModel) GetUserByID(id primitive.ObjectID) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = model.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

// GetUserByViolet 通过 VioletID 查找用户
func (model *UserModel) GetUserByViolet(id string) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = model.Collection.FindOne(ctx, bson.M{"violet_id": id}).Decode(&user)
	return
}

// GetUserByWechat 通过 Wechat OpenID 查找用户
func (model *UserModel) GetUserByWechat(id string) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = model.Collection.FindOne(ctx, bson.M{"wechat_id": id}).Decode(&user)
	return
}

// SetUserInfoByID 更新用户个人信息
func (model *UserModel) SetUserInfoByID(id primitive.ObjectID, info UserInfoSchema) error {
	ctx, over := GetCtx()
	defer over()
	// 通过反射获取非空字段
	updateItem := bson.M{}
	names := reflect.TypeOf(info)
	values := reflect.ValueOf(info)
	for i := 0; i < names.NumField(); i++ {
		name := names.Field(i).Tag.Get("bson")
		if name == "birthday" { // 生日字段为 int64
			if values.Field(i).Int() != 0 {
				updateItem["info."+name] = values.Field(i).Int()
			}
		} else { // 其他字段为 string
			if values.Field(i).String() != "" {
				updateItem["info."+name] = values.Field(i).String()
			}
		}
	}
	if res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateItem}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	// 更新缓存
	return GetRedis().Cache.WillUpdate(id, KindOfBaseInfo)
}

// UserDataCount 用户数据更新
type UserDataCount struct {
	Money           int64 `bson:"money"`             // 当前持有闲币
	Value           int64 `bson:"value"`             // 用户积分
	Credit          int64 `bson:"credit"`            // 个人信誉
	PublishCount    int64 `bson:"publish_count"`     // 发布任务数
	PublishRunCount int64 `bson:"publish_run_count"` // 发布并进行中任务数
	ReceiveCount    int64 `bson:"receive_count"`     // 领取任务数
	ReceiveRunCount int64 `bson:"receive_run_count"` // 领取并进行中任务数
	FollowingCount  int64 `bson:"following_count"`   // 关注人数量
	FollowerCount   int64 `bson:"follower_count"`    // 粉丝数量
}

// UpdateUserDataCount 更新用户数值数据（偏移值）
func (model *UserModel) UpdateUserDataCount(id primitive.ObjectID, data UserDataCount) error {
	ctx, over := GetCtx()
	defer over()
	// 通过反射获取非零字段
	updateItem := bson.M{}
	names := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < names.NumField(); i++ {
		if value := values.Field(i).Int(); value != 0 {
			updateItem["data."+names.Field(i).Tag.Get("bson")] = value
		}
	}
	if res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$inc": updateItem}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// SetUserType 设置用户类型
func (model *UserModel) SetUserType(id primitive.ObjectID, userType UserType) error {
	ctx, over := GetCtx()
	defer over()
	if res, err := model.Collection.UpdateMany(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"data.type": userType}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	// 更新缓存
	return GetRedis().Cache.WillUpdate(id, KindOfBaseInfo)
}

// SetUserAttend 用户签到
func (model *UserModel) SetUserAttend(id primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	if res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"data.attendance_date": time.Now().Unix()}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// SetUserCertification 设置用户认证信息
func (model *UserModel) SetUserCertification(id primitive.ObjectID, data UserCertificationSchema) error {
	ctx, over := GetCtx()
	defer over()
	if res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"certification": data}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// CheckCertificationEmail 是否存在邮箱认证
func (model *UserModel) CheckCertificationEmail(email string) bool {
	ctx, over := GetCtx()
	defer over()
	if res, err := model.Collection.CountDocuments(ctx,
		bson.M{"certification.email": email, "certification.status": CertificationTrue}); err == nil && res == 0 {
		return true
	}
	return false
}

// AddCollectTask 添加任务收藏
func (model *UserModel) AddCollectTask(id, taskID primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	res, err := model.Collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$addToSet": bson.M{string("collect_tasks"): taskID}})
	if err != nil {
		return err
	} else if res.UpsertedCount == 0 && res.ModifiedCount == 0 {
		return errors.New("exist")
	}
	return nil
}

// RemoveCollectTask 移除任务收藏
func (model *UserModel) RemoveCollectTask(id, taskID primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	res, err := model.Collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$pull": bson.M{string("collect_tasks"): taskID}})
	if err != nil {
		return err
	} else if res.UpsertedCount == 0 && res.ModifiedCount == 0 {
		return errors.New("exist")
	}
	return nil
}

// TODO 添加关注

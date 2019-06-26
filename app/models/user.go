package models

import (
	"errors"
	"github.com/TimeForCoin/Server/app/utils"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

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

// CertificationStatus 用户认证状态
const (
	CertificationNone       CertificationStatus = "none"        // 未认证
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
	SearchHistory  []string             `bson:"search_history"`  // 搜索历史(仅保留最近的 20 条)
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

func makeNewUserSchema() UserSchema {
	return UserSchema{
		RegisterTime: time.Now().Unix(),
		Data: UserDataSchema{
			Type:          UserTypeNormal,
			CollectTasks:  []primitive.ObjectID{},
			SearchHistory: []string{},
			Money: 100,
			Value: 1000,
			Credit: 100,
		},
		Certification: UserCertificationSchema{
			Identity: IdentityNone,
			Status: CertificationNone,
		},
	}
}

// AddUserByViolet 通过 Violet 增加用户
func (m *UserModel) AddUserByViolet(id string) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()
	userID := primitive.NewObjectID()
	newUser := makeNewUserSchema()
	newUser.ID = userID
	newUser.VioletID = id
	_, err := m.Collection.InsertOne(ctx, newUser)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return userID, nil
}

// AddUserByWechat 通过微信添加用户
func (m *UserModel) AddUserByWechat(openid string) (primitive.ObjectID, error) {
	ctx, finish := GetCtx()
	defer finish()
	userID := primitive.NewObjectID()
	newUser := makeNewUserSchema()
	newUser.ID = userID
	newUser.WechatID = openid
	newUser.Info.Nickname = "微信用户" + utils.GetRandomString(6)
	newUser.Info.Avatar = "https://coin-1252808268.cos.ap-guangzhou.myqcloud.com/avatar-5cfe5cab2cfbe5ed600f9665.png"
	_, err := m.Collection.InsertOne(ctx, newUser)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return userID, nil
}

// GetUserByID 通过 ID 查找用户
func (m *UserModel) GetUserByID(id primitive.ObjectID) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = m.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

// GetUserByViolet 通过 VioletID 查找用户
func (m *UserModel) GetUserByViolet(id string) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = m.Collection.FindOne(ctx, bson.M{"violet_id": id}).Decode(&user)
	return
}

// GetUserByWechat 通过 Wechat OpenID 查找用户
func (m *UserModel) GetUserByWechat(id string) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = m.Collection.FindOne(ctx, bson.M{"wechat_id": id}).Decode(&user)
	return
}

// SetUserInfoByID 更新用户个人信息
func (m *UserModel) SetUserInfoByID(id primitive.ObjectID, info UserInfoSchema) error {
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
	if res, err := m.Collection.UpdateOne(ctx,
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
func (m *UserModel) UpdateUserDataCount(id primitive.ObjectID, data UserDataCount) error {
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
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$inc": updateItem}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// GetUsers 按关键字搜索用户
func (m *UserModel) GetUsers(key string, page, size int64) (res []UserSchema) {
	ctx, over := GetCtx()
	defer over()
	cursor, err := m.Collection.Find(ctx, bson.M{"info.nickname": bson.M{"$regex": key, "$options": "$i"}},
		options.Find().SetSkip((page-1)*size).SetLimit(size))
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		user := UserSchema{}
		err = cursor.Decode(&user)
		if err != nil {
			return
		}
		res = append(res, user)
	}
	return
}

// SetUserType 设置用户类型
func (m *UserModel) SetUserType(id primitive.ObjectID, userType UserType) error {
	ctx, over := GetCtx()
	defer over()
	if res, err := m.Collection.UpdateMany(ctx,
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
func (m *UserModel) SetUserAttend(id primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"data.attendance_date": time.Now().Unix()}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// SetUserCertification 设置用户认证信息
func (m *UserModel) SetUserCertification(id primitive.ObjectID, data UserCertificationSchema) error {
	ctx, over := GetCtx()
	defer over()
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"certification": data}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// GetCertification 获取认证
func (m *UserModel) GetCertification(status []CertificationStatus, page, size int64) (res []UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	cur, err := m.Collection.Find(ctx, bson.M{"certification.status": bson.M{"$in": status}}, options.Find().SetSkip((page - 1) * size).SetLimit(size))
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		user := UserSchema{}
		err = cur.Decode(&user)
		if err != nil {
			return
		}
		res = append(res, user)
	}
	return
}

// CheckCertificationEmail 该邮箱是否存在认证
func (m *UserModel) CheckCertificationEmail(email string) bool {
	ctx, over := GetCtx()
	defer over()
	if res, err := m.Collection.CountDocuments(ctx,
		bson.M{"certification.email": email, "certification.status": CertificationTrue}); err == nil && res == 0 {
		return true
	}
	return false
}

// AddCollectTask 添加任务收藏
func (m *UserModel) AddCollectTask(id, taskID primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$addToSet": bson.M{"data.collect_tasks": taskID}})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return errors.New("exist")
	}
	return nil
}

// RemoveCollectTask 移除任务收藏
func (m *UserModel) RemoveCollectTask(id, taskID primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$pull": bson.M{"data.collect_tasks": taskID}})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return ErrNotExist
	}
	return nil
}

// AddSearchHistory 添加搜索历史
func (m *UserModel) AddSearchHistory(id primitive.ObjectID, key string) error {
	ctx, over := GetCtx()
	defer over()
	res, _ := m.Collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$addToSet": bson.M{"data.search_history": key}})
	if res.MatchedCount == 0 {
		return ErrNotExist
	}
	return nil
}

// ClearSearchHistory 清空搜索历史
func (m *UserModel) ClearSearchHistory(id primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()
	res, _ := m.Collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{"data.search_history": []string{}}})
	if res.MatchedCount == 0 {
		return ErrNotExist
	}
	return nil
}

// GetSearchHistory 获取用户搜索历史
func (m *UserModel) GetSearchHistory(id primitive.ObjectID) ([]string, error) {
	var user UserSchema
	ctx, over := GetCtx()
	defer over()
	err := m.Collection.FindOne(ctx, bson.M{"_id": id},
		options.FindOne().SetProjection(bson.M{"data.search_history": 1})).Decode(&user)
	if err != nil {
		return []string{}, ErrNotExist
	}
	return user.Data.SearchHistory, nil
}

// GetAllUser 获取所有用户ID
func (m *UserModel) GetAllUser() (res []primitive.ObjectID) {
	ctx, over := GetCtx()
	defer over()
	cur, err := m.Collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		user := UserSchema{}
		err = cur.Decode(&user)
		if err != nil {
			return
		}
		res = append(res, user.ID)
	}
	return res
}

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
)

// UserSchema User 基本数据结构
type UserSchema struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"` // 用户ID [索引]
	OpenID   string             `bson:"open_id"`       // 微信OpenID
	VioletID string             `bson:"violet_id"`     // VioletID
	Name     string             `bson:"name"`          // 用户名， 唯一
	// 用户个性信息
	Info struct {
		Email    string     // 联系邮箱
		Phone    string     // 联系手机
		Avatar   string     // 头像
		Nickname string     // 用户昵称
		Bio      string     // 个人简介
		School   string     // 学校
		Gender   UserGender // 性别
		Location string     // 具体位置
	}
	// 用户数据
	Data struct {
		Money          int64                // 当前持有闲币
		Exp            int64                // 等级经验
		Value          int64                // 用户积分
		Credit         int64                // 个人信誉
		Type           UserType             // 用户类型
		AttendanceDate int                  // 签到时间戳
		CollectTasks   []primitive.ObjectID `bson:"collect_tasks"` // 收藏的任务
		// 冗余数据
		PublishFinishTasks int64 `bson:"publish_finish_tasks"` // 发布并完成任务数
		ReceiveTasks       int64 `bson:"receive_tasks"`        // 领取任务数
		ReceiveFinishTasks int64 `bson:"receive_finish_tasks"` // 领取并完成任务数
	}
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

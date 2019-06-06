package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemModel 系统 数据库
type SystemModel struct {
	Collection *mongo.Collection
}

// SystemSchemas 系统数据
// 存储一些系统信息，管理信息 键值对
type SystemSchemas struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"` // ID
	Key    string             // 信息名称
	Value  string             // 信息内容
}

func (m *SystemModel) ExistAutoEmail(email string) string {
	ctx, finish := GetCtx()
	defer finish()
	data := SystemSchemas{}
	 err := m.Collection.FindOne(ctx, bson.M{"key":"email-"+email}).Decode(&data)
	if err != nil {
		return ""
	}
	return data.Value
}

func (m *SystemModel) AddAutoEmail(email, data string) error {
	ctx, finish := GetCtx()
	defer finish()
	opt := options.Update()
	opt.SetUpsert(true)
	_, err := m.Collection.UpdateOne(ctx, bson.M{"key":"email-"+email}, bson.M{"$set": bson.M{"value": data}}, opt)
	return err
}

// 获取/设置 自动通过验证的邮箱后缀名

// 获取/设置 平均用户信用水平

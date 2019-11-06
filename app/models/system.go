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
	ID    primitive.ObjectID `bson:"_id,omitempty"` // ID
	Key   string             // 信息名称
	Value string             // 信息内容
}

// GetAutoEmail 获取自动认证后缀
func (m *SystemModel) GetAutoEmail(page, size int64) (res []SystemSchemas, err error) {
	ctx, finish := GetCtx()
	defer finish()
	cur, err := m.Collection.Find(ctx, bson.M{"key": bson.M{"$regex": "^email-"}}, options.Find().SetSkip((page-1)*size).SetLimit(size))
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		item := SystemSchemas{}
		err = cur.Decode(&item)
		if err != nil {
			return
		}
		res = append(res, item)
	}
	return
}

// ExistAutoEmail 是否存在自动认证
func (m *SystemModel) ExistAutoEmail(email string) string {
	ctx, finish := GetCtx()
	defer finish()
	data := SystemSchemas{}
	err := m.Collection.FindOne(ctx, bson.M{"key": "email-" + email}).Decode(&data)
	if err != nil {
		return ""
	}
	return data.Value
}

// AddAutoEmail 添加自动认证
func (m *SystemModel) AddAutoEmail(email, data string) error {
	ctx, finish := GetCtx()
	defer finish()
	opt := options.Update()
	opt.SetUpsert(true)
	_, err := m.Collection.UpdateOne(ctx, bson.M{"key": "email-" + email}, bson.M{"$set": bson.M{"value": data}}, opt)
	return err
}

// RemoveAutoEmail 移除自动认证
func (m *SystemModel) RemoveAutoEmail(email string) error {
	ctx, finish := GetCtx()
	defer finish()
	res, err := m.Collection.DeleteOne(ctx, bson.M{"key": "email-" + email})
	if err != nil {
		return err
	} else if res.DeletedCount == 0 {
		return ErrNotExist
	}
	return nil
}

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SystemModel 系统 数据库
type SystemModel struct {
	Collection *mongo.Collection
}

// SystemSchemas 系统数据
// 存储一些系统信息，管理信息
type SystemSchemas struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"` // ID
	Key    string             // 信息名称
	Value  string             // 信息内容
	Values struct {           // 变量信息
		Key   string
		Value string
	}
}

// 获取/设置 自动通过验证的邮箱后缀名

// 获取/设置 平均用户信用水平

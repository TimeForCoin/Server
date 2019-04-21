package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RelationModel 文件数据库
type RelationModel struct {
	Collection *mongo.Collection
}

// RelationSchemas 关系数据结构
type RelationSchemas struct {
	UserID    primitive.ObjectID   `bson:"_id"` // 用户ID
	Following []primitive.ObjectID // 关注的人
	Follower  []primitive.ObjectID // 粉丝
}

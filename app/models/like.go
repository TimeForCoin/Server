package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// LikeModel 点赞数据库
type LikeModel struct {
	Collection *mongo.Collection
}

// LikeSchemas 点赞数据结构
type LikeSchemas struct {
	ContentID primitive.ObjectID   `bson:"_id"`     // 内容ID
	LikeID    []primitive.ObjectID `bson:"like_id"` // 点赞的ID
}

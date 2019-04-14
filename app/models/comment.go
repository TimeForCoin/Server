package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CommentModel 评论数据库
type CommentModel struct {
	Collection *mongo.Collection
}

// CommentSchema 评论数据结构
type CommentSchema struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"` // 评论 ID
	ContentID  primitive.ObjectID   `bson:"content_id"`    // 被评论内容 ID [索引]
	ContentOwn primitive.ObjectID   `bson:"content_own"`   // 被评论内容 用户 ID
	UserID     primitive.ObjectID   `bson:"user_id"`       // 评论用户 ID
	ReplyCount int64                `bson:"reply_count"`   // 回复数
	LikeCount  int64                `bson:"like_count"`    // 点赞数(冗余)
	LikeID     []primitive.ObjectID `bson:"like_id"`       // 点赞用户ID
	Content    string               `bson:"content"`       // 评论内容
	Time       int64                `bson:"time"`          // 评论时间
}

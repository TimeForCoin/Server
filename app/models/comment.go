package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	ContentOwn primitive.ObjectID   `bson:"content_own"`   // 被评论内容 用户 ID(冗余)
	UserID     primitive.ObjectID   `bson:"user_id"`       // 评论用户 ID
	ReplyCount int64                `bson:"reply_count"`   // 回复数
	LikeCount  int64                `bson:"like_count"`    // 点赞数(冗余)
	Content    string               `bson:"content"`       // 评论内容
	Time       int64                `bson:"time"`          // 评论时间
}

// 添加评论
func (m *CommentModel) AddComment(contentID, contentOwn, userID primitive.ObjectID, content string) error {
	ctx, finish := GetCtx()
	defer finish()
	_, err := m.Collection.InsertOne(ctx, &CommentSchema{
		ContentID:  contentID,
		ContentOwn: contentOwn,
		UserID:     userID,
		Content:    content,
		Time:       time.Now().Unix(),
	})
	return err
}

// 获取评论
func (m *CommentModel) GetCommentsByContent(contentID primitive.ObjectID) (res []CommentSchema, err error) {
	ctx, finish := GetCtx()
	defer finish()
	cur, err := m.Collection.Find(ctx, bson.M{"content_id": contentID})
	if err != nil {
		return
	}
	//noinspection GoUnhandledErrorResult
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result CommentSchema
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		res = append(res, result)
	}
	err = cur.Err()
	return
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

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
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`        // 评论 ID
	ContentID  primitive.ObjectID `bson:"content_id" json:"content_id"`   // 被评论内容 ID [索引]
	ContentOwn primitive.ObjectID `bson:"content_own" json:"content_own"` // 被评论内容 用户 ID(冗余)
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`         // 评论用户 ID
	ReplyCount int64              `bson:"reply_count"`                    // 回复数
	LikeCount  int64              `bson:"like_count"`                     // 点赞数(冗余)
	Content    string             `bson:"content"`                        // 评论内容
	Time       int64              `bson:"time"`                           // 评论时间
	IsDelete   bool               `bson:"is_delete"`                      // 是否已被删除
	IsReply    bool               `bson:"id_reply" json:"is_reply"`       // 是否为回复
}

// AddComment 添加评论
func (m *CommentModel) AddComment(contentID, contentOwn, userID primitive.ObjectID, content string, isReply bool) error {
	ctx, finish := GetCtx()
	defer finish()
	_, err := m.Collection.InsertOne(ctx, &CommentSchema{
		ContentID:  contentID,
		ContentOwn: contentOwn,
		UserID:     userID,
		Content:    content,
		Time:       time.Now().Unix(),
		IsDelete:   false,
		IsReply:    isReply,
	})
	return err
}

// GetCommentsByContent 分页获取评论
func (m *CommentModel) GetCommentsByContent(contentID primitive.ObjectID, page, size int64, sort bson.M) (res []CommentSchema, err error) {
	ctx, finish := GetCtx()
	defer finish()
	cur, err := m.Collection.Find(ctx, bson.M{"content_id": contentID},
		options.Find().SetSkip((page-1)*size).SetLimit(size).SetSort(sort))
	if err != nil {
		return
	}
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

// GetCommentByID 获取指定评论
func (m *CommentModel) GetCommentByID(commentID primitive.ObjectID) (res CommentSchema, err error) {
	ctx, finish := GetCtx()
	defer finish()
	err = m.Collection.FindOne(ctx, bson.M{"_id": commentID}).Decode(&res)
	return
}

// RemoveContentByID 删除评论
func (m *CommentModel) RemoveContentByID(commentID primitive.ObjectID) error {
	ctx, finish := GetCtx()
	defer finish()
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": commentID},
		bson.M{"$set": bson.M{"is_delete": true, "content": ""}})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return ErrNotExist
	}
	return nil
}

// InsertCount 增加计数
func (m *CommentModel) InsertCount(commentID primitive.ObjectID, name ContentCountType, count int64) error {
	ctx, finish := GetCtx()
	defer finish()
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": commentID},
		bson.M{"$inc": bson.M{string(name): count}})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return ErrNotExist
	}
	return nil
}

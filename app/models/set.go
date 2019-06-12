package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetModel 集合数据库
type SetModel struct {
	Collection *mongo.Collection
}

// SetSchemas 集合数据结构
type SetSchemas struct {
	UserID        primitive.ObjectID   `bson:"_id"`             // 用户ID
	LikeTaskID    []primitive.ObjectID `bson:"like_task_id"`    // 点赞的任务ID
	LikeCommentID []primitive.ObjectID `bson:"like_comment_id"` // 点赞的评论ID
	CollectTaskID []primitive.ObjectID `bson:"collect_task_id"` // 收藏的任务ID
}

// SetKind 集合类型
type SetKind string

// SetKind 集合类型
const (
	SetOfLikeTask    SetKind = "like_task_id"
	SetOfLikeComment SetKind = "like_comment_id"
	SetOfCollectTask SetKind = "collect_task_id"
)

// GetSets 获取集合
func (m *SetModel) GetSets(userID primitive.ObjectID, kind SetKind) (set SetSchemas) {
	ctx, finish := GetCtx()
	defer finish()
	opt := options.FindOne()
	opt.SetProjection(bson.M{string(kind): 1})
	_ = m.Collection.FindOne(ctx, bson.M{"_id": userID}, opt).Decode(&set)
	return
}

// AddToSet 添加到集合
func (m *SetModel) AddToSet(userID, targetID primitive.ObjectID, kind SetKind) error {
	ctx, finish := GetCtx()
	defer finish()
	opt := options.Update()
	opt.SetUpsert(true)
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": userID},
		bson.M{"$addToSet": bson.M{string(kind): targetID}}, opt)
	if err != nil {
		return err
	} else if res.UpsertedCount == 0 && res.ModifiedCount == 0 {
		return ErrNotExist
	}
	return nil
}

// RemoveFromSet 从集合中移除
func (m *SetModel) RemoveFromSet(userID, targetID primitive.ObjectID, kind SetKind) error {
	ctx, finish := GetCtx()
	defer finish()
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": userID},
		bson.M{"$pull": bson.M{string(kind): targetID}})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return ErrNotExist
	}
	return nil
}

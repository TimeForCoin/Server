package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

// MessageModel 消息数据库
type MessageModel struct {
	Collection *mongo.Collection
}

// MessageType 消息类型
type MessageType string

// MessageType 消息类型
const (
	MessageTypeChat    MessageType = "chat"    // 聊天
	MessageTypeSystem  MessageType = "system"  // 系统通知
	MessageTypeTask    MessageType = "task"    // 任务通知
	MessageTypeComment MessageType = "comment" // 评论通知
)

// MessageSchema 消息数据结构
type MessageSchema struct {
	UserID  primitive.ObjectID `bson:"user" json:"user_id"`            // 消息发言人ID
	Time    int64              `bson:"time"`            // 发送时间
	Title   string             `bson:"title,omitempty"` // 消息标题 (系统通知/任务通知)
	Content string             `bson:"content"`         // 消息内容
	About   primitive.ObjectID `bson:"about,omitempty"` // 相关ID (被评论的任务)
}

// SessionSchema Session 数据结构
// bson 默认为名字小写
type SessionSchema struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`  // 消息会话ID
	User1       primitive.ObjectID `bson:"user_1" json:"user_1"`     // 用户1 [索引]
	User2       primitive.ObjectID `bson:"user_2" json:"user_2"`     // 用户2/任务ID [索引]
	Unread1     int64              `bson:"unread_1" json:"unread_1"` // 用户1 未读消息数量
	Unread2     int64              `bson:"unread_2" json:"unread_2"` // 用户2 未读消息数量
	Type        MessageType        `bson:"type"`                     // 会话类型
	LastMessage MessageSchema      `bson:"last_message"`             // 最新的消息[冗余]
	Messages    []MessageSchema    `bson:"messages" json:"messages"` // 消息内容
}

// GetSessionsByUser 获取会话列表
func (m *MessageModel) GetSessionsByUser(userID primitive.ObjectID, page, size int64) (res []SessionSchema) {
	ctx, over := GetCtx()
	defer over()
	cur, err := m.Collection.Find(ctx, bson.M{"$or": []bson.M{{"user_1": userID}, {"user_2": userID}}},
		options.Find().
			SetProjection(bson.M{"messages": 0}).
			SetSkip((page-1)*size).SetLimit(size).
			SetSort(bson.M{"last_message.time": -1}))
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		session := SessionSchema{}
		err = cur.Decode(&session)
		if err != nil {
			return
		}
		res = append(res, session)
	}
	return
}

// GetSessionByID 获取指定会话详情
func (m *MessageModel) GetSessionByID(id primitive.ObjectID) (res SessionSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = m.Collection.FindOne(ctx, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"messages": 0})).Decode(&res)
	return
}

// GetSessionWithMsgByID 获取会话详情(附带信息)
func (m *MessageModel) GetSessionWithMsgByID(id primitive.ObjectID, page, size int64) (res SessionSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = m.Collection.FindOne(ctx, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{
		"messages": bson.M{"$slice": []int64{(page - 1) * size, page * size}},
	})).Decode(&res)
	return
}

// AddMessage 添加信息
func (m *MessageModel) AddMessage(recUser primitive.ObjectID, messageType MessageType, data MessageSchema) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()
	// 固定顺序
	unread := bson.M{}
	user1, user2 := data.UserID, recUser
	if strings.Compare(user1.Hex(), user2.Hex()) < 0 {
		user1, user2 = user2, user1
		unread["unread_1"] = 1
	} else {
		unread["unread_2"] = 1
	}
	data.Time = time.Now().Unix()
	var res SessionSchema
	err := m.Collection.FindOneAndUpdate(ctx, bson.M{
		"user_1": user1, "user_2": user2, "type": messageType,
	}, bson.M{
		"$push": bson.M{
			"messages": bson.M{
				"$each": []MessageSchema{data},
				"$sort": bson.M{"time": -1},
			},
		},
		"$set": bson.M{"last_message": data},
		"$inc": unread,
	}, options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After).
		SetProjection(bson.M{"_id": 1})).Decode(&res)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.ID, nil
}

// ReadMessage 标记信息为已读
func (m *MessageModel) ReadMessage(sessionID primitive.ObjectID, firstUser bool) error {
	ctx, over := GetCtx()
	defer over()
	unread := bson.M{}
	if firstUser {
		unread["unread_1"] = 0
	} else {
		unread["unread_2"] = 0
	}
	res, err := m.Collection.UpdateOne(ctx, bson.M{
		"_id": sessionID,
	}, bson.M{
		"$set": unread,
	})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return ErrNotExist
	}
	return nil
}

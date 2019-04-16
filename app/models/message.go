package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

// MessageSchema Message 数据结构
// bson 默认为名字小写
type MessageSchema struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"` // 消息会话ID
	User1    primitive.ObjectID `bson:"user_1"`        // 用户1 [索引]
	User2    primitive.ObjectID `bson:"user_2"`        // 用户2/任务ID [索引]
	Unread1  int64              `bson:"unread_1"`      // 用户1 未读消息数量
	Unread2  int64              `bson:"unread_2"`      // 用户2 未读消息数量
	LastTime int64              `bson:"last_time"`     // 最新消息时间(冗余)
	Type     MessageType        `bson:"type"`          // 会话类型
	Messages []struct {
		Time    int64              `bson:"time"`             // 发送时间
		Title   string             `bson:"title, omitempty"` // 消息标题 (系统通知/任务通知)
		Content string             `bson:"content"`          // 消息内容
		About   primitive.ObjectID `bson:"about, omitempty"` // 相关ID (被评论的任务)
	}
}

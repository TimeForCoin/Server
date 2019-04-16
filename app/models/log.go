package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// LogModel 日志数据库
type LogModel struct {
	Collection *mongo.Collection
}

// LogType 日志类型
type LogType string

// LogType 日志类型
const (
	LogTypeMoney  LogType = "user_money"   // 金钱变动
	LogTypeValue  LogType = "user_value"   // 用户积分变动
	LogTypeCredit LogType = "user_credit"  // 信用变动
	LogTypeLogin  LogType = "user_login"   // 用户登陆
	LogTypeClear  LogType = "system_clear" // 系统清理
	LogTypeStart  LogType = "system_start" // 系统启动
	LogTypeError  LogType = "system_error" // 系统运行时错误
)

// LogSchema 日志数据结构
type LogSchema struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"` // 日志ID
	Time    int64              // 时间
	Type    LogType            // 日志类型
	UserID  primitive.ObjectID `bson:"user_id"`            // 相关用户 [索引]
	AboutID primitive.ObjectID `bson:"about_id,omitempty"` // 相关事件
	Value   int64              `bson:"value,omitempty"`    // 数值
	Msg     string             `bson:"msg,omitempty"`      // 消息
}

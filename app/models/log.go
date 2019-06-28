package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (m *LogModel) AddLog(userID, aboutID primitive.ObjectID, logType LogType) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()

	logID := primitive.NewObjectID()
	_, err := m.Collection.InsertOne(ctx, &LogSchema{
		ID:      logID,
		Time:    time.Now().Unix(),
		Type:    logType,
		UserID:  userID,
		AboutID: aboutID,
	})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return logID, nil
}

func (m *LogModel) SetValue(id primitive.ObjectID, value int64) error {
	ctx, over := GetCtx()
	defer over()

	if res, err := m.Collection.UpdateMany(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"value": value}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

func (m *LogModel) SetMsg(id primitive.ObjectID, msg string) error {
	ctx, over := GetCtx()
	defer over()

	if res, err := m.Collection.UpdateMany(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"msg": msg}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

func (m *LogModel) GetLog(userID primitive.ObjectID, logTypes []LogType,
	startDate, endDate int64, skip, limit int64) (logs []LogSchema, count int64, err error) {
	ctx, over := GetCtx()
	defer over()

	filter := bson.M{
		"type":    bson.M{"$in": logTypes},
		"user_id": userID,
		"time":    bson.M{"$gt": startDate},
	}

	if endDate != 0 {
		filter["time"] = bson.M{"$lt": endDate, "$gt": startDate}
	}

	count, err = m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return
	}

	cursor, err := m.Collection.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(limit))
	if err != nil {
		return
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		log := LogSchema{}
		err = cursor.Decode(&log)
		if err != nil {
			return
		}
		logs = append(logs, log)
	}

	return
}

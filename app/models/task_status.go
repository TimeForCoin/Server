package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskStatusModel 接受的任务状态数据库
type TaskStatusModel struct {
	Collection *mongo.Collection
}

// PlayerStatus 参与用户状态
type PlayerStatus string

// PlayerStatus 参与用户状态
const (
	PlayerWait    PlayerStatus = "wait"    // 等待同意加入
	PlayerRefuse  PlayerStatus = "refuse"  // 拒绝加入
	PlayerClose   PlayerStatus = "close"   // 发布者关闭任务
	PlayerRunning PlayerStatus = "running" // 用户进行中
	PlayerFinish  PlayerStatus = "finish"  // 用户已完成
	PlayerGiveUp  PlayerStatus = "give_up" // 用户已放弃
	PlayerFailure PlayerStatus = "failure" // 任务失败
)

// TaskStatusSchema 接受的任务状态 基本数据结构
// bson 默认为名字小写
type TaskStatusSchema struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"` // 任务状态ID
	Task   primitive.ObjectID `bson:"task"`          // 任务 ID [索引]
	Player primitive.ObjectID `bson:"player"`        // 用户 ID [索引]
	Status PlayerStatus       // 状态
	Note   string             // 申请备注
	// 完成后的评价
	Degree int    // 完成度
	Remark string // 评语
	// 用户的反馈
	Score    int    // 五星好评
	Feedback string // 反馈
}

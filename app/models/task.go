package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskModel Task 任务 数据库
type TaskModel struct {
	Collection *mongo.Collection
}

// TaskType 任务类型
type TaskType string

// RewardType 酬劳类型
type RewardType string

// TaskStatus 任务状态
type TaskStatus string

// AttachType 附件类型
type AttachType string

// TaskType 任务类型
const (
	TaskTypeRunning       TaskType = "run"           // 跑腿任务
	TaskTypeQuestionnaire TaskType = "questionnaire" // 问卷任务
	TaskTypeInfo          TaskType = "info"          // 信息任务
)

// RewardType 酬劳类型
const (
	RewardMoney    RewardType = "money"     // 闲钱币酬劳
	RewardRMB      RewardType = "rmb"       // 人民币酬劳
	RewardPhysical RewardType = "physical " // 实物酬劳
)

// TaskStatus 任务状态
const (
	TaskStatusDraft   TaskStatus = "draft"   // 草稿
	TaskStatusWait    TaskStatus = "wait"    // 等待接受
	TaskStatusClose   TaskStatus = "close"   // 已关闭
	TaskStatusFinish  TaskStatus = "finish"  // 已完成
	TaskStatusOverdue TaskStatus = "overdue" // 已过期
)

// AttachType 附件类型
const (
	AttachImage AttachType = "image" // 图片
	AttachFile  AttachType = "file"  // 文件
)

// AttachmentSchema 附件数据结构
// 附件存储在本地文件系统中 ./{任务ID}/{附件ID}
type AttachmentSchema struct {
	ID       primitive.ObjectID // 附件ID
	Type     AttachType         // 附件类型
	Name     string             // 附件名
	Describe string             // 附件描述
	Size     int64              // 附件大小
	Time     int64              // 创建时间
	Use      bool               // 是否使用，未使用附件将定期处理
}

// TaskSchema Task 基本数据结构
type TaskSchema struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // 任务ID
	Publisher primitive.ObjectID `bson:"publisher"`     // 任务发布者 [索引]

	Title      string             // 任务名称
	Type       TaskType           // 任务类型
	Content    string             // 任务内容
	Attachment []AttachmentSchema // 任务附件
	Status     TaskStatus         // 任务状态

	Reward       RewardType // 酬劳类型
	RewardValue  int        `bson:"reward_value, omitempty"`  // 酬劳数值
	RewardObject string     `bson:"reward_object, omitempty"` // 酬劳物体

	PublishDate int64 `bson:"publish_date"` // 任务发布时间
	EndDate     int64 `bson:"end_data"`     // 任务结束时间

	Player    int64 `bson:"player"`     // 参与的用户
	MaxPlayer int64 `bson:"max_player"` // 参与用户上限

	ViewCount    int64                `bson:"view_count"`    // 任务浏览数
	CollectCount int64                `bson:"collect_count"` // 收藏数
	CommentCount int64                `bson:"comment_count"` // 评论数(冗余)
	LikeCount    int64                `bson:"like_count"`    // 点赞数(冗余)
	LikeID       []primitive.ObjectID `bson:"like_id"`       // 点赞用户ID
}

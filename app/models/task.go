package models

import (
	"reflect"
	"time"
	"go.mongodb.org/mongo-driver/bson"
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

// TaskType 任务类型
const (
	TaskTypeRunning       TaskType = "run"           // 跑腿任务
	TaskTypeQuestionnaire TaskType = "questionnaire" // 问卷任务
	TaskTypeInfo          TaskType = "info"          // 信息任务
)

// RewardType 酬劳类型
const (
	RewardMoney    RewardType = "money"    // 闲钱币酬劳
	RewardRMB      RewardType = "rmb"      // 人民币酬劳
	RewardPhysical RewardType = "physical" // 实物酬劳
)

// TaskStatus 任务状态
const (
	TaskStatusDraft   TaskStatus = "draft"   // 草稿
	TaskStatusWait    TaskStatus = "wait"    // 等待接受
	TaskStatusRun     TaskStatus = "run"     // 执行中（人数已满）
	TaskStatusClose   TaskStatus = "close"   // 已关闭
	TaskStatusFinish  TaskStatus = "finish"  // 已完成
	TaskStatusOverdue TaskStatus = "overdue" // 已过期
)

// TaskSchema Task 基本数据结构
type TaskSchema struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // 任务ID
	Publisher primitive.ObjectID `bson:"publisher"`     // 任务发布者 [索引]

	Title      string               `bson:"title"`// 任务名称
	Type       TaskType             `bson:"type"`// 任务类型
	Content    string               `bson:"content"`// 任务内容
	Attachment []primitive.ObjectID `bson:"attachment"`// 任务附件
	Status     TaskStatus           `bson:"status"`// 任务状态
	Location   []string             `bson:"location"`// 任务地点 (非问卷类任务)
	Tags       []string             `bson:"tags"`// 标签 (作为关键词，改进搜索体验)
	TopTime    int64                `bson:"top_time"`// 置顶时间(默认为0)，如果当前时间小于置顶时间，即将任务置顶

	Reward       RewardType `bson:"reward"`// 酬劳类型
	RewardValue  float32    `bson:"reward_value, omitempty"`  // 酬劳数值
	RewardObject string     `bson:"reward_object, omitempty"` // 酬劳物体

	PublishDate int64 `bson:"publish_date"` // 任务发布时间
	StartDate   int64 `bson:"start_date"`   // 任务开始时间
	EndDate     int64 `bson:"end_date"`     // 任务结束时间

	PlayerCount int64 `bson:"player_count"` // 参与的用户
	MaxPlayer   int64 `bson:"max_player"`   // 参与用户上限, -1为无限制
	MaxFinish   int64 `bson:"max_finish"`   // 完成用户上限, 可用于收集指定数量问卷
	AutoAccept  bool  `bson:"auto_accept"`  // 自动同意领取任务

	ViewCount    int64                `bson:"view_count"`    // 任务浏览数
	CollectCount int64                `bson:"collect_count"` // 收藏数
	CommentCount int64                `bson:"comment_count"` // 评论数(冗余)
	LikeCount    int64                `bson:"like_count"`    // 点赞数(冗余)
	LikeID       []primitive.ObjectID `bson:"like_id"`       // 点赞用户ID

	// 由[浏览量、评论数、收藏数、参与人数、时间、置顶、酬劳、发布者粉丝、信用]等数据加权计算，10分钟更新一次，用于排序
	Hot int64 `bson:"hot"` // 任务热度
}

func (model *TaskModel) AddTask(publisherID string) (string, error) {
	ctx, over := GetCtx()
	defer over()
	taskID := primitive.NewObjectID()
	pid, err := primitive.ObjectIDFromHex(publisherID)
	_, err = model.Collection.InsertOne(ctx, &TaskSchema{
		ID:          taskID,
		Publisher:   pid,
		PublishDate: time.Now().Unix(),
	})
	if err != nil {
		return "", err
	}
	return taskID.Hex(), nil
}

func (model *TaskModel) SetTaskInfoByID(id string, info TaskSchema) error {
	ctx, over := GetCtx()
	defer over()
	_id, err := primitive.ObjectIDFromHex(id)
	//_uid, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		return err
	}
	// 通过反射获取非空字段
	updateItem := bson.M{}
	names := reflect.TypeOf(info)
	values := reflect.ValueOf(info)
	for i := 0; i < names.NumField(); i++ {
		name := names.Field(i).Tag.Get("bson")
		if name == "top_time" || name == "publish_date" || name == "start_date" || name == "end_date" || name == "player_count" || name == "max_player" || name == "max_finish" ||  name == "view_count" || name == "collect_count" || name == "comment_count" || name == "like_count" || name == "hot"{ // 生日字段为 int64
			if values.Field(i).Int() != 0 {
				updateItem[name] = values.Field(i).Int()
			}
		} else if name == "reward_value" {
			if values.Field(i).Float() != 0 {
				updateItem[name] = values.Field(i).Float()
			}
		} else if name == "auto_accept" {
			updateItem[name] = values.Field(i).Bool()
		} else if name == "attachment" || name == "location" || name == "tags" || name == "like_id" {
			if values.Field(i).Len() != 0 {
				//len := values.Field(i).Len()
				//updateItem[name] = values.Field(i).Slice(0, len)
			}
		} else if name == "publisher" {
			
		}else { // 其他字段为 string
			if values.Field(i).String() != "" {
				updateItem[name] = values.Field(i).String()
			}
		}
	}
	//updateItem["publisher"] = _uid
	if res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": _id},
		bson.M{"$set": updateItem}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	// 更新缓存
	return GetRedis().Cache.WillUpdateBaseInfo(id)
}
func (model *TaskModel) GetTaskByID(id string) (task TaskSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	err = model.Collection.FindOne(ctx, bson.M{"_id": _id}).Decode(&task)
	return
}
package models

import (
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	RewardMoney  RewardType = "money"  // 闲钱币酬劳
	RewardRMB    RewardType = "rmb"    // 人民币酬劳
	RewardObject RewardType = "object" // 实物酬劳
)

// TaskStatus 任务状态
const (
	TaskStatusDraft  TaskStatus = "draft"  // 草稿
	TaskStatusWait   TaskStatus = "wait"   // 等待接受
	TaskStatusClose  TaskStatus = "close"  // 已关闭
	TaskStatusFinish TaskStatus = "finish" // 已完成
)

// TaskSchema Task 基本数据结构
type TaskSchema struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"` // 任务ID
	Publisher primitive.ObjectID `bson:"publisher"`               // 任务发布者 [索引]

	Title    string     `bson:"title"`    // 任务名称
	Type     TaskType   `bson:"type"`     // 任务类型
	Content  string     `bson:"content"`  // 任务内容
	Status   TaskStatus `bson:"status"`   // 任务状态
	Location []string   `bson:"location"` // 任务地点 (非问卷类任务)
	Tags     []string   `bson:"tags"`     // 标签 (作为关键词，改进搜索体验)
	TopTime  int64      `bson:"top_time"` // 置顶时间(默认为0)，如果当前时间小于置顶时间，即将任务置顶

	Reward       RewardType `bson:"reward"`        // 酬劳类型
	RewardValue  float32    `bson:"reward_value"`  // 酬劳数值
	RewardObject string     `bson:"reward_object"` // 酬劳物体

	PublishDate int64 `bson:"publish_date"` // 任务发布时间
	StartDate   int64 `bson:"start_date"`   // 任务开始时间
	EndDate     int64 `bson:"end_date"`     // 任务结束时间

	PlayerCount int64 `bson:"player_count"` // 参与的用户
	MaxPlayer   int64 `bson:"max_player"`   // 参与用户上限, -1为无限制
	AutoAccept  bool  `bson:"auto_accept"`  // 自动同意领取任务

	ViewCount    int64 `bson:"view_count"`    // 任务浏览数
	CollectCount int64 `bson:"collect_count"` // 收藏数
	CommentCount int64 `bson:"comment_count"` // 评论数(冗余)
	LikeCount    int64 `bson:"like_count"`    // 点赞数(冗余)

	// 由[浏览量、评论数、收藏数、参与人数、时间、置顶、酬劳、发布者粉丝、信用]等数据加权计算，10分钟更新一次，用于排序
	Hot int64 `bson:"hot"` // 任务热度
}

func (m *TaskModel) AddTask(taskID, publisherID primitive.ObjectID, status TaskStatus) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()
	res, err := m.Collection.InsertOne(ctx, &TaskSchema{
		ID: taskID,
		Publisher:   publisherID,
		PublishDate: time.Now().Unix(),
		Status:      status,
	})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (m *TaskModel) SetTaskInfoByID(id primitive.ObjectID, info TaskSchema) error {
	ctx, over := GetCtx()
	defer over()
	// 通过反射获取非空字段
	updateItem := bson.M{}
	names := reflect.TypeOf(info)
	values := reflect.ValueOf(info)
	for i := 0; i < names.NumField(); i++ {
		name := names.Field(i).Tag.Get("bson")
		if name == "top_time" || name == "publish_date" || name == "start_date" || name == "end_date" || name == "player_count" || name == "max_player" {
			if values.Field(i).Int() != 0 {
				updateItem[name] = values.Field(i).Int()
			}
		} else if name == "reward_value" {
			if values.Field(i).Float() != 0 {
				updateItem[name] = values.Field(i).Float()
			}
		} else if name == "auto_accept" {
			updateItem[name] = values.Field(i).Bool()
		} else if name == "title" || name == "type" || name == "content" || name == "reward" || name == "reward_object" || name == "status" { // 其他字段为 string
			if values.Field(i).String() != "" {
				updateItem[name] = values.Field(i).String()
			}
		}
	}
	if len(info.Location) > 0 {
		updateItem["location"] = info.Location
	}
	if len(info.Tags) > 0 {
		updateItem["tags"] = info.Tags
	}
	//updateItem["publisher"] = _uid
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateItem}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	// 更新缓存
	return nil
}
func (m *TaskModel) GetTaskByID(id primitive.ObjectID) (task TaskSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = m.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	return
}

//type TaskCard struct {
//	ID           string  // 任务ID
//	Publisher    string  // 发布者ID
//	Avatar       string  // 发布者头像
//	Credit       int64   // 发布者信用度
//	Title        string  // 任务标题
//	TopTime      int64   `json:"top_time"` // 任务是否置顶
//	EndDate      int64   `json:"end_date"` // 任务截止时间
//	Reward       string  // 酬劳类型
//	RewardValue  float32 `json:"reward_value, omitempty"`  // 酬劳数值
//	RewardObject string  `json:"reward_object, omitempty"` // 酬劳物体
//}

// 获取任务列表，需要按类型/状态/酬劳类型筛选，按关键词搜索，按不同规则排序
func (m *TaskModel) GetTasks(sort string, taskIDs []primitive.ObjectID, taskTypes []TaskType,
	statuses []TaskStatus, rewards []RewardType, keywords []string, user string, skip, limit int64) (tasks []TaskSchema, count int64, err error) {
	ctx, over := GetCtx()
	defer over()

	var keywordsRegex string
	for i, str := range keywords {
		keywordsRegex += "(" + str + ")"
		if i != len(keywords)-1 {
			keywordsRegex += "|"
		}
	}

	// TODO 关键词筛选
	// 按类型、状态、酬劳类型、关键词筛选
	filter := bson.M{}
	if len(taskIDs) > 0 {
		filter = bson.M{
			"_id":    bson.M{"$in": taskIDs},
			"type":   bson.M{"$in": taskTypes},
			"status": bson.M{"$in": statuses},
			//"tags":    bson.M{"$in": keywords},
			"reward": bson.M{"$in": rewards}}
	} else {
		filter = bson.M{
			"type":   bson.M{"$in": taskTypes},
			"status": bson.M{"$in": statuses},
			//"tags":    bson.M{"$in": keywords},
			"reward": bson.M{"$in": rewards}}
	}
	//"title":   bson.M{"$regex": keywordsRegex},
	//"content": bson.M{"$regex": keywordsRegex}}

	// 筛选发布者
	if user != "" {
		var _id primitive.ObjectID
		_id, err = primitive.ObjectIDFromHex(user)
		if err != nil {
			return
		}
		filter["publisher"] = _id
	}

	count, err = m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return
	}

	cursor, err := m.Collection.Find(ctx, filter, options.Find().SetSort(bson.M{sort: -1}).SetSkip(skip).SetLimit(limit))
	if err != nil {
		return
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		task := TaskSchema{}
		err = cursor.Decode(&task)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

func (m *TaskModel) RemoveTask(taskID primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()

	res, err := m.Collection.DeleteOne(ctx, bson.M{"_id": taskID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotExist
	}
	return nil
}

type ContentCountType string

const (
	ViewCount    ContentCountType = "view_count"    // 任务浏览数
	CollectCount ContentCountType = "collect_count" // 收藏数
	CommentCount ContentCountType = "comment_count" // 评论数(冗余)
	LikeCount    ContentCountType = "like_count"    // 点赞数(冗余)
	ReplyCount   ContentCountType = "reply_count"   // 回复数
)

func (m *TaskModel) InsertCount(taskID primitive.ObjectID, name ContentCountType, count int) error {
	ctx, over := GetCtx()
	defer over()
	res, err := m.Collection.UpdateOne(ctx, bson.M{"_id": taskID}, bson.M{"$inc": bson.M{string(name): count}})
	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return ErrNotExist
	}

	return nil
}

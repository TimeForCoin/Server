package models

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"` // 任务状态ID
	Task   primitive.ObjectID `bson:"task"`                    // 任务 ID [索引]
	Player primitive.ObjectID `bson:"player"`                  // 用户 ID [索引]
	Status PlayerStatus       `bson:"status"`                  // 状态
	Note   string             `bson:"note"`                    // 申请备注
	// 完成后的评价
	Degree int    `bson:"degree"` // 完成度
	Remark string `bson:"remark"` // 评语
	// 用户的反馈
	Score    int    `bson:"score"`    // 五星好评
	Feedback string `bson:"feedback"` // 反馈
}

// AddTaskStatus 添加任务状态
func (m *TaskStatusModel) AddTaskStatus(taskStatusID, taskID, userID primitive.ObjectID, status PlayerStatus) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()
	res, err := m.Collection.InsertOne(ctx, &TaskStatusSchema{
		ID:     taskStatusID,
		Task:   taskID,
		Player: userID,
		Status: status,
	})
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

// SetTaskStatus 设置任务状态
func (m *TaskStatusModel) SetTaskStatus(id primitive.ObjectID, info TaskStatusSchema) error {
	ctx, over := GetCtx()
	defer over()
	// 通过反射获取非空字段
	updateItem := bson.M{}
	names := reflect.TypeOf(info)
	values := reflect.ValueOf(info)
	for i := 0; i < names.NumField(); i++ {
		name := names.Field(i).Tag.Get("bson")
		if name == "degree" || name == "score" {
			if values.Field(i).Int() != 0 {
				updateItem[name] = values.Field(i).Int()
			}
		} else if name == "status" || name == "note" || name == "remark" || name == "feedback" { // 其他字段为 string
			if values.Field(i).String() != "" {
				updateItem[name] = values.Field(i).String()
			}
		}
	}

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

// GetTaskStatusListByTaskID 获取任务状态列表
func (m *TaskStatusModel) GetTaskStatusListByTaskID(taskID primitive.ObjectID, status []PlayerStatus, skip, limit int64) (taskStatusList []TaskStatusSchema, count int64, err error) {
	ctx, over := GetCtx()
	defer over()

	filter := bson.M{
		"task":   taskID,
		"status": bson.M{"$in": status},
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
		taskStatus := TaskStatusSchema{}
		err = cursor.Decode(&taskStatus)
		if err != nil {
			return
		}
		taskStatusList = append(taskStatusList, taskStatus)
	}

	return
}

// GetTaskStatusListByUserID 获取用户任务状态列表
func (m *TaskStatusModel) GetTaskStatusListByUserID(userID primitive.ObjectID, status []PlayerStatus, skip, limit int64) (taskStatusList []TaskStatusSchema, count int64, err error) {
	ctx, over := GetCtx()
	defer over()

	filter := bson.M{
		"player": userID,
		"status": bson.M{"$in": status},
	}
	fmt.Println(filter)

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
		taskStatus := TaskStatusSchema{}
		err = cursor.Decode(&taskStatus)
		if err != nil {
			return
		}
		taskStatusList = append(taskStatusList, taskStatus)
	}

	return
}

// GetTaskStatus 获取任务状态
func (m *TaskStatusModel) GetTaskStatus(userID, taskID primitive.ObjectID) (taskStatus TaskStatusSchema, err error) {
	ctx, over := GetCtx()
	defer over()

	filter := bson.M{
		"task":   taskID,
		"player": userID,
	}

	err = m.Collection.FindOne(ctx, filter).Decode(&taskStatus)
	return
}

// DeleteTaskStatus 删除任务状态
func (m *TaskStatusModel) DeleteTaskStatus(id primitive.ObjectID) error {
	ctx, over := GetCtx()
	defer over()

	res, err := m.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotExist
	}
	return nil
}

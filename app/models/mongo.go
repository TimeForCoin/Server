package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/TimeForCoin/Server/app/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var model *Model

// ErrNotExist 数据不存在
var ErrNotExist = errors.New("not_exist")

// Model 数据库实例
type Model struct {
	client *mongo.Client
	db     *mongo.Database
	// 数据库实例
	Log           *LogModel
	Article       *ArticleModel
	Comment       *CommentModel
	Message       *MessageModel
	Questionnaire *QuestionnaireModel
	Task          *TaskModel
	TaskStatus    *TaskStatusModel
	User          *UserModel
	File          *FileModel
	Set           *SetModel
	System        *SystemModel
}

// GetModel 获取 Model 实例
func GetModel() *Model {
	return model
}

// GetCtx 获取并发上下文(默认10秒超时)
func GetCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// createIndexes 检查并创建索引
func createIndexes(ctx context.Context, name string, indexes []bson.M) error {
	collection := model.db.Collection(name)
	collectionIndexes := collection.Indexes()
	cur, err := collectionIndexes.List(ctx)
	if err != nil { // 读取索引发生错误
		return err
	}
	if cur == nil { // 指针不存在
		return errors.New("can't read collection")
	}
	if !cur.Next(ctx) { // 索引不存在，创建索引
		log.Info().Msg("Init index for " + name)
		for i := range indexes { // 创建唯一索引
			if _, err := collectionIndexes.CreateOne(ctx, mongo.IndexModel{
				Keys:    indexes[i],
				Options: options.Index().SetUnique(false),
			}); err != nil {
				return err
			}
		}
	}
	return cur.Close(ctx) // 关闭指针
}

// initCollection 初始化集合
func initCollection() error {
	ctx, cancel := GetCtx()
	defer cancel()
	// 初始化索引
	DBIndexes := []struct {
		name    string
		indexes []bson.M
	}{
		{name: "comments", indexes: []bson.M{{"content_id": 1}}},
		{name: "messages", indexes: []bson.M{{"user_1": 1}, {"user_2": 1}}},
		{name: "tasks", indexes: []bson.M{{"publisher": 1}}},
		{name: "logs", indexes: []bson.M{{"user_id": 1}}},
		{name: "task_status", indexes: []bson.M{{"task": 1}, {"player": 1}}},
		{name: "files", indexes: []bson.M{{"owner_id": 1}}},
	}
	for _, i := range DBIndexes {
		if err := createIndexes(ctx, i.name, i.indexes); err != nil {
			return err
		}
	}
	return nil
}

// InitDB 初始化数据库
func InitDB(config *utils.DBConfig) error {
	model = &Model{}
	err := connect(config)
	if err != nil {
		return err
	}
	model.db = model.client.Database(config.DBName)

	// 初始化集合
	if err := initCollection(); err != nil {
		return err
	}
	// 初始化 Model
	// 公告数据库
	model.Article = &ArticleModel{
		Collection: model.db.Collection("article"),
	}
	// 日志数据库
	model.Log = &LogModel{
		Collection: model.db.Collection("logs"),
	}
	// 问卷数据库
	model.Questionnaire = &QuestionnaireModel{
		Collection: model.db.Collection("questionnaires"),
	}
	// 评论数据库
	model.Comment = &CommentModel{
		Collection: model.db.Collection("comments"),
	}
	// 消息数据库
	model.Message = &MessageModel{
		Collection: model.db.Collection("messages"),
	}
	// 任务状态数据库
	model.TaskStatus = &TaskStatusModel{
		Collection: model.db.Collection("task_status"),
	}
	// 用户数据库
	model.User = &UserModel{
		Collection: model.db.Collection("users"),
	}
	// 任务数据库
	model.Task = &TaskModel{
		Collection: model.db.Collection("tasks"),
	}
	// 文件数据库
	model.File = &FileModel{
		Collection: model.db.Collection("files"),
	}
	// 点赞数据库
	model.Set = &SetModel{
		Collection: model.db.Collection("sets"),
	}
	// 系统数据库
	model.System = &SystemModel{
		Collection: model.db.Collection("system"),
	}
	return nil
}

// 连接数据库
func connect(config *utils.DBConfig) error {
	ctx, cancel := GetCtx()
	defer cancel()
	option := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			config.User, config.Password, config.Host, config.Port, config.DBName))
	var err error
	if model.client, err = mongo.Connect(ctx, option); err != nil {
		return err
	}
	// 测试连接
	if err := model.client.Ping(ctx, readpref.Primary()); err != nil {
		log.Error().Err(err).Msg("Failure to connect MongoDB!!!")
		return err
	}
	log.Info().Msg("Successful connection to MongoDB.")
	return nil
}

// DisconnectDB 断开数据库连接
func DisconnectDB() error {
	if model == nil {
		return nil
	}
	ctx, cancel := GetCtx()
	defer cancel()
	err := model.client.Disconnect(ctx)
	model = nil
	return err
}

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/TimeForCoin/Server/app/configs"
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
	Comment       *CommentModel
	Message       *MessageModel
	Questionnaire *QuestionnaireModel
	Task          *TaskModel
	TaskStatus    *TaskStatusModel
	User          *UserModel
}

// GetModel 获取 Model 实例
func GetModel() *Model {
	if model == nil {
		panic("DB isn't Initialize!")
	}
	return model
}

// GetCtx 获取并发上下文(默认10秒超时)
func GetCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// createIndexes 检查并创建索引
func createIndexes(ctx context.Context, collection string, indexes []bson.M) error {
	questionnaires := model.db.Collection(collection)
	questionnairesIndexes := questionnaires.Indexes()
	cur, err := questionnairesIndexes.List(ctx)
	if err != nil {
		return err
	}
	if cur == nil {
		return errors.New("can't read collection")
	}
	if !cur.Next(ctx) {
		// 创建索引
		for i := range indexes {
			indexOptions := options.Index()
			indexOptions.SetUnique(true)
			if _, err := questionnairesIndexes.CreateOne(ctx, mongo.IndexModel{
				Keys:    indexes[i],
				Options: indexOptions,
			}); err != nil {
				return err
			}
		}
	}
	if err := cur.Close(ctx); err != nil {
		return err
	}
	return nil
}

// initCollection 初始化索引
func initCollection() error {
	ctx, cancel := GetCtx()
	defer cancel()
	if err := createIndexes(ctx, "comments", []bson.M{{"content_id": 1}}); err != nil {
		return err
	}
	if err := createIndexes(ctx, "messages", []bson.M{{"user_1": 1}, {"user_2": 1}}); err != nil {
		return err
	}
	// 任务数据库
	if err := createIndexes(ctx, "tasks", []bson.M{{"publisher": 1}}); err != nil {
		return err
	}
	if err := createIndexes(ctx, "questionnaires", []bson.M{{"task_id": 1}}); err != nil {
		return err
	}
	if err := createIndexes(ctx, "task_status", []bson.M{{"task": 1, "owner": 1}}); err != nil {
		return err
	}
	return nil

}

// InitDB 初始化数据库
func InitDB(config *configs.DBConfig) error {
	model = &Model{}
	err := connect(config)
	if err != nil {
		return err
	}
	model.db = model.client.Database(config.DBName)

	// 初始化 索引
	if err := initCollection(); err != nil {
		return err
	}
	// 初始化 Model
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

	return nil
}

// 连接数据库
func connect(config *configs.DBConfig) error {
	var err error
	ctx, cancel := GetCtx()
	defer cancel()
	option := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			config.User, config.Password, config.Host, config.Port, config.DBName))
	if model.client, err = mongo.Connect(ctx, option); err != nil {
		return err
	}
	// 测试连接
	err = model.client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error().Err(err).Msg("Failure to connect MongoDB!!!")
		return err
	}
	log.Info().Msg("Successful connection to MongoDB.")
	return nil
}

// DisconnectDB 断开数据库连接
func DisconnectDB() error {
	ctx, cancel := GetCtx()
	defer cancel()
	return model.client.Disconnect(ctx)
}

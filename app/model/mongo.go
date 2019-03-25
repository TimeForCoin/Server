package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/TimeForCoin/Server/app/configs"
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
	User   *UserModel
}

// GetModel 获取 Model 实例
func GetModel() *Model {
	if model == nil {
		panic("DB isn't Initialize!")
	}
	return model
}

// GetCtx 获取并发上下文(默认3秒超时)
func GetCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

// InitDB 初始化数据库
func InitDB(config *configs.DBConfig) error {
	model = &Model{}
	err := connect(config)
	if err != nil {
		return err
	}
	// 初始化 Collection 实例
	model.db = model.client.Database(config.DBName)
	model.User = &UserModel{
		Collection: model.db.Collection("users"),
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
		log.Println("Failure to connect MongoDB!!!")
		return err
	}
	log.Println("Successful connection to MongoDB.")
	return nil
}

// DisconnectDB 断开数据库连接
func DisconnectDB() error {
	ctx, cancel := GetCtx()
	defer cancel()
	return model.client.Disconnect(ctx)
}

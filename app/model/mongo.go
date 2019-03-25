package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/TimeForCoin/Server/app/configs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

var model *Model

var ErrNotExist = errors.New("not_exist")

type Model struct {
	client *mongo.Client
	db     *mongo.Database
	User   *UserModel
}

// GetCtx 获取并发上下文(默认3秒超时)
func GetCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

// InitDB 初始化数据库
func InitDB(config configs.DBConfig) error {
	model = &Model{}
	err := connectDB(config)
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
func connectDB(config configs.DBConfig) error {
	var err error
	ctx, cancel := GetCtx()
	defer cancel()
	option := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			config.User, config.Password, config.Host, config.Port, config.DBName))
	model.client, err = mongo.Connect(ctx, option)
	if err != nil {
		return err
	}
	// 测试连接
	err = model.client.Ping(ctx, readpref.Primary())
	if err == nil {
		fmt.Println("MongoDB connected " + config.Host)
	} else {
		fmt.Println("MongoDB connected failed.")
		return err
	}
	return nil
}

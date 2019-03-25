package model

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserModel User 数据库
type UserModel struct {
	Collection *mongo.Collection
}

// UserSchema User 基本数据结构
type UserSchema struct {
	ID   primitive.ObjectID "_id,omitempty"
	Name string             `bson:"name"`
}

func (model *UserModel) AddUser(name string) error {
	ctx, over := GetCtx()
	defer over()
	// 返回ID
	_, err := model.Collection.InsertOne(ctx, &UserSchema{Name: name})
	return err
}

func (model *UserModel) FindUser(name string) (user UserSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = model.Collection.FindOne(ctx, bson.M{"name": name}).Decode(&user)
	return
}

func (model *UserModel) UpdateUser(name string, newName string) error {
	ctx, over := GetCtx()
	defer over()
	// 返回更新的数量
	res, err := model.Collection.UpdateOne(ctx, bson.M{"name": name}, bson.M{"$set": bson.M{"name": newName}})
	if err != nil {
		return nil
	}
	if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

func (model *UserModel) RemoveUser(name string) error {
	ctx, over := GetCtx()
	defer over()
	res, err := model.Collection.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return ErrNotExist
	}
	return nil
}

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ArticleModel 文章数据库
type ArticleModel struct {
	Collection *mongo.Collection
}

// ArticleSchemas 文章数据结构
type ArticleSchemas struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // ID
	ViewCount int64              `bson:"view_count"`    // 文章阅读数
	Title     string             // 文章标题
	Content   string             // 文章内容
	Publisher string             // 发布者名字
	Date      int64              // 发布时间
	Image     primitive.ObjectID // 首页图片
}

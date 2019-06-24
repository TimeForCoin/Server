package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// ArticleModel 公告文章数据库
type ArticleModel struct {
	Collection *mongo.Collection
}

// ArticleSchemas 公告文章数据结构
type ArticleSchemas struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // ID
	ViewCount int64              `bson:"view_count"`    // 文章阅读数
	Title     string             // 文章标题
	Content   string             // 文章内容
	Publisher string             // 发布者名字
	Date      int64              // 发布时间
	Images     []primitive.ObjectID // 首页图片
}

// GetArticles 获取公告列表
func (m *ArticleModel) GetArticles(skip, limit int64) (articles []ArticleSchemas, count int64, err error){
	ctx, over := GetCtx()
	defer over()

	filter := bson.M{}
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
		article := ArticleSchemas{}
		err = cursor.Decode(&article)
		if err != nil {
			return
		}
		articles = append(articles, article)
	}
	return
}

// AddArticle 添加公告文章
func (m *ArticleModel) AddArticle(articleID primitive.ObjectID, title string, content string, publisher string, images []primitive.ObjectID) (id primitive.ObjectID, err error) {
	ctx, over := GetCtx()
	defer over()

	res, err := m.Collection.InsertOne(ctx, &ArticleSchemas{
		ID:			articleID,
		ViewCount:	0,
		Title:		title,
		Content:	content,
		Publisher:	publisher,
		Date:		time.Now().Unix(),
		Images:		images,
	})

	if err != nil {
		return primitive.ObjectID{}, err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

// GetArticleByID 根据ID获取公告文章详情
func (m *ArticleModel) GetArticleByID(id primitive.ObjectID) (article ArticleSchemas, err error) {
	ctx, over := GetCtx()
	defer over()

	err = m.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	return
}

// SetArticleByID 根据ID修改公告文章
func (m *ArticleModel) SetArticleByID(id primitive.ObjectID, title, content, publisher string, images []primitive.ObjectID) (err error) {
	ctx, over := GetCtx()
	defer over()

	updateItem := bson.M{
		"title":		title,
		"content":		content,
		"publisher":	publisher,
		"images":		images,
	}
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateItem}); err != nil {
			return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

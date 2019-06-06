package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// FileModel 文件数据库
type FileModel struct {
	Collection *mongo.Collection
}

// FileType 文件类型
type FileType string

// OwnerType 文件归属类型
type OwnerType string

// FileType 文件类型
const (
	FileImage FileType = "image" // 图片
	FileFile  FileType = "file"  // 文件
)

// OwnerType 文件归属类型
const (
	FileForUser OwnerType = "user" // 用户文件，非公开内容仅用户本人查看[认证材料、问卷/数据征集提交内容]
	FileForTask OwnerType = "task" // 任务文件，非公开内容仅任务参与者查看[任务附件]
)

// FileSchema 文件数据结构
type FileSchema struct {
	ID          primitive.ObjectID // 文件ID[索引]
	OwnerID     primitive.ObjectID `bson:"owner_id"` // 拥有者ID[索引]
	Owner       OwnerType          // 问卷归属类型
	Type        FileType           // 文件类型
	Name        string             // 文件名
	Description string             // 文件描述
	Size        int64              // 文件大小
	Time        int64              // 创建时间
	Use         int64              // 引用数，未使用文件将定期处理
	Public      bool               // 公开，非公开文件需要验证权限
}
func (model *FileModel) GetFileByID(id string) (file FileSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	err = model.Collection.FindOne(ctx, bson.M{"_id": _id}).Decode(&file)
	return
}
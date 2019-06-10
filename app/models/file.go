package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	FileForTask OwnerType = "task" // 任务文件，非公开内容仅任务参与者查看[任务附件/图片]
)

// FileSchema 文件数据结构
type FileSchema struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // 文件ID[索引]
	Time        int64              // 创建时间
	Used        int                // 引用数
	OwnerID     primitive.ObjectID `bson:"owner_id"` // 拥有者ID[索引]
	Owner       OwnerType          // 文件归属类型
	Type        FileType           // 文件类型
	Name        string             // 文件名
	Description string             // 文件描述
	Size        int64              // 文件大小
	Public      bool               // 公开，非公开文件需要验证权限
	URL         string             // 下载链接
}

// 添加文件
func (m *FileModel) AddFile(fileID, ownerID primitive.ObjectID, owner OwnerType, fileType FileType,
	name, description, url string, size int64, public, used bool) error {
	ctx, finish := GetCtx()
	defer finish()
	usedNum := 0
	if used {
		usedNum = 1
	}
	_, err := m.Collection.InsertOne(ctx, &FileSchema{
		ID:          fileID,
		Time:        time.Now().Unix(),
		OwnerID:     ownerID,
		Owner:       owner,
		Type:        fileType,
		Name:        name,
		Description: description,
		Size:        size,
		Public:      public,
		Used:        usedNum,
		URL:         url,
	})
	fmt.Println("err", err)
	if err != nil {
		return  err
	}
	return nil
}

// 获取文件信息
func (m *FileModel) GetFile(id primitive.ObjectID) (res FileSchema, err error) {
	ctx, finish := GetCtx()
	defer finish()
	err = m.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	return
}

// 获取内容文件
func (m *FileModel) GetFileByContent(id primitive.ObjectID, fileType ...FileType) (res []FileSchema, err error) {
	ctx, finish := GetCtx()
	defer finish()

	search := bson.M{"owner_id": id}
	if len(fileType) > 0 {
		search["type"] = fileType[0]
	}

	cur, err := m.Collection.Find(ctx, search)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result FileSchema
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		res = append(res, result)
	}
	err = cur.Err()
	return
}

// 将文件绑定到任务中
func (m *FileModel) BindTask(fileID, taskID primitive.ObjectID) error {
	ctx, finish := GetCtx()
	defer finish()
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": fileID},
		bson.M{"$set": bson.M{"owner_id": taskID, "owner": FileForTask}, "$inc": bson.M{"used": 1}}); err != nil {
		return err
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

func (m *FileModel) RemoveFile(fileID primitive.ObjectID) error {
	ctx, finish := GetCtx()
	defer finish()
	if res, err := m.Collection.DeleteOne(ctx,
		bson.M{"_id": fileID}); err != nil {
		return err
	} else if res.DeletedCount < 1 {
		return ErrNotExist
	}
	return nil
}

func (m *FileModel) SetFileInfo(fileID primitive.ObjectID, name, description string, public bool) error {
	ctx, finish := GetCtx()
	defer finish()
	if res, err := m.Collection.UpdateOne(ctx,
		bson.M{"_id": fileID},
		bson.M{"$set": bson.M{"name": name, "description": description, "public": public}}); err != nil {
		return err
	} else if res.ModifiedCount < 1 {
		return ErrNotExist
	}
	return nil
}
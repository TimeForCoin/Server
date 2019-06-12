package services

import (
	"mime/multipart"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileService 用户逻辑
type FileService interface {
	AddFile(file multipart.File, head multipart.FileHeader, fileType models.FileType,
		ownerType models.OwnerType, ownID primitive.ObjectID,
		name, description string, public bool) primitive.ObjectID
	BindFilesToTask(userID, taskID primitive.ObjectID, files []FileBaseInfo)
	RemoveFiles(fileID primitive.ObjectID)
	UpdateFileInfo(fileID, userID primitive.ObjectID, name, description string, public bool)
}

// NewUserService 初始化
func newFileService() FileService {
	return &fileService{
		model:     models.GetModel().File,
		taskModel: models.GetModel().Task,
	}
}

type fileService struct {
	model     *models.FileModel
	taskModel *models.TaskModel
}

// AddFile 添加文件
func (s *fileService) AddFile(file multipart.File, head multipart.FileHeader, fileType models.FileType,
	ownerType models.OwnerType, ownID primitive.ObjectID,
	name, description string, public bool) primitive.ObjectID {

	fileID := primitive.NewObjectID()
	// 上传到腾讯云
	url, err := libs.GetCOS().SaveFile("file-"+fileID.Hex()+name, file)
	libs.AssertErr(err, "", 500)
	// 保存到数据库
	err = s.model.AddFile(fileID, ownID, ownerType, fileType, name, description, url, head.Size, public, false)
	libs.AssertErr(err, "", 500)
	return fileID
}

// FileBaseInfo 文件基本信息
type FileBaseInfo struct {
	ID   primitive.ObjectID
	Type models.FileType
}

// BindFilesToTask 添加文件到任务中
func (s *fileService) BindFilesToTask(userID, taskID primitive.ObjectID, files []FileBaseInfo) {
	// 验证权限
	for _, file := range files {
		f, err := s.model.GetFile(file.ID)
		libs.AssertErr(err, "faked_file", 403)
		libs.Assert(f.OwnerID == userID, "permission_deny", 403)
		libs.Assert(f.Type == file.Type, "error_file_type", 403)
	}
	for _, file := range files {
		err := s.model.BindTask(file.ID, taskID)
		libs.AssertErr(err, "", 500)
	}
}

// RemoveFiles 移除文件
func (s *fileService) RemoveFiles(fileID primitive.ObjectID) {
	f, err := s.model.GetFile(fileID)
	libs.AssertErr(err, "", 500)
	err = s.model.RemoveFile(fileID)
	libs.AssertErr(err, "", 500)
	err = libs.GetCOS().DeleteFile(fileID.Hex() + f.Name)
	libs.AssertErr(err, "", 500)
}

// UpdateFileInfo 更新文件信息
func (s *fileService) UpdateFileInfo(fileID, userID primitive.ObjectID, name, description string, public bool) {
	file, err := s.model.GetFile(fileID)
	libs.AssertErr(err, "faked_file", 403)
	// 检验权限
	if file.Owner == models.FileForUser {
		libs.Assert(file.OwnerID == userID, "permission_deny")
	} else if file.Owner == models.FileForTask {
		task, err := s.taskModel.GetTaskByID(file.OwnerID)
		libs.AssertErr(err, "", 500)
		libs.Assert(task.Publisher == userID, "permission_deny")
	}
	err = s.model.SetFileInfo(fileID, name, description, public)
	libs.AssertErr(err, "", 500)
}

package services

import (
	"github.com/TimeForCoin/Server/app/utils"
	"mime/multipart"
	"path"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileService 用户逻辑
type FileService interface {
	AddFile(file multipart.File, head multipart.FileHeader, fileType models.FileType,
		ownID primitive.ObjectID, name, description string, public bool) primitive.ObjectID
	BindFilesToTask(userID, taskID primitive.ObjectID, files []FileBaseInfo)
	BindFilesToUser(userID primitive.ObjectID, files []primitive.ObjectID)
	RemoveFile(fileID primitive.ObjectID)
	UpdateFileInfo(fileID, userID primitive.ObjectID, name, description string, public bool)
	RemoveUserFile(userID, fileID primitive.ObjectID)
	RemoveUselessFile(userID primitive.ObjectID, all bool) (removeCount int64)
}

// newFileService 初始化
func newFileService() FileService {
	return &fileService{
		model:     models.GetModel().File,
		taskModel: models.GetModel().Task,
		cache:     models.GetRedis().Cache,
	}
}

type fileService struct {
	model     *models.FileModel
	taskModel *models.TaskModel
	cache     *models.CacheModel
}

// AddFile 添加文件
func (s *fileService) AddFile(file multipart.File, head multipart.FileHeader, fileType models.FileType, ownID primitive.ObjectID,
	name, description string, public bool) primitive.ObjectID {

	fileID := primitive.NewObjectID()
	fileHash, err := utils.GetFileHash(file)
	utils.AssertErr(err, "", 500)
	// 寻找相同文件
	sameFile, err := s.model.GetFileByHash(fileHash)
	var url string
	var cosName string
	if err == nil && sameFile.Size == head.Size {
		// 存在相同的图片
		url = sameFile.URL
		cosName = sameFile.COSName
	} else {
		// 上传到腾讯云
		cosName = "file-" + fileID.Hex() + "-" + fileHash + path.Ext(head.Filename)
		url, err = libs.GetCOS().SaveFile(cosName, file)
		utils.AssertErr(err, "", 500)
	}

	// 默认公开图片
	if fileType == models.FileImage {
		public = true
	}

	// 保存到数据库
	err = s.model.AddFile(models.FileSchema{
		ID:          fileID,
		OwnerID:     ownID,
		Owner:       models.FileForUser,
		Type:        fileType,
		Name:        name,
		Description: description,
		URL:         url,
		Size:        head.Size,
		Public:      public,
		Used:        0,
		Hash:        fileHash,
		COSName:     cosName,
	})
	utils.AssertErr(err, "", 500)
	return fileID
}

// FileBaseInfo 文件基本信息
type FileBaseInfo struct {
	ID   primitive.ObjectID
	Type models.FileType
}

func (s *fileService) GetFiles() {
	// TODO 验证权限获取文件列表
}

// BindFilesToTask 添加文件到任务中
func (s *fileService) BindFilesToTask(userID, taskID primitive.ObjectID, files []FileBaseInfo) {
	// 验证权限
	for _, file := range files {
		f, err := s.model.GetFile(file.ID)
		utils.AssertErr(err, "faked_file", 403)
		utils.Assert(f.OwnerID == userID, "permission_deny", 403)
		utils.Assert(f.Type == file.Type, "error_file_type", 403)
	}
	for _, file := range files {
		err := s.model.BindTask(file.ID, taskID)
		utils.AssertErr(err, "", 500)
	}
}

// BindFilesToUser 绑定文件到用户上
func (s *fileService) BindFilesToUser(userID primitive.ObjectID, files []primitive.ObjectID) {
	// 验证权限
	for _, file := range files {
		f, err := s.model.GetFile(file)
		utils.AssertErr(err, "faked_file", 403)
		utils.Assert(f.OwnerID == userID, "permission_deny", 403)
		utils.Assert(f.Owner == models.FileForUser, "permission_deny", 403)
	}
	for _, file := range files {
		err := s.model.BindUser(file)
		utils.AssertErr(err, "", 500)
	}
}

// RemoveFiles 移除文件
func (s *fileService) RemoveFile(fileID primitive.ObjectID) {
	f, err := s.model.GetFile(fileID)
	utils.AssertErr(err, "", 500)
	err = s.model.RemoveFile(fileID)
	utils.AssertErr(err, "", 500)
	err = libs.GetCOS().DeleteFile(f.COSName)
	utils.AssertErr(err, "", 500)
}

// RemoveFiles 移除无用文件
func (s *fileService) RemoveUselessFile(userID primitive.ObjectID, all bool) (removeCount int64) {
	// 验证权限
	var files []models.FileSchema
	if all {
		user, err := s.cache.GetUserBaseInfo(userID)
		utils.AssertErr(err, "", 500)
		utils.Assert(user.Type == models.UserTypeAdmin || user.Type == models.UserTypeRoot, "permission_deny", 403)
		files = s.model.GetUselessFile()
		removeCount = s.model.RemoveUselessFile()
	} else {
		files = s.model.GetUselessFile(userID)
		removeCount = s.model.RemoveUselessFile(userID)
	}

	for _, file := range files {
		_, err := s.model.GetFileByHash(file.Hash)
		if err != nil {
			// 没有相同的文件
			err = libs.GetCOS().DeleteFile(file.COSName)
			utils.AssertErr(err, "", 500)
		}
	}
	return
}

// RemoveUserFile 移除用户临时文件
func (s *fileService) RemoveUserFile(userID, fileID primitive.ObjectID) {
	f, err := s.model.GetFile(fileID)
	utils.AssertErr(err, "faked_file", 403)
	if f.Owner == models.FileForUser {
		utils.Assert(f.OwnerID == userID, "permission_deny", 403)
	} else if f.Owner == models.FileForTask {
		task, err := s.taskModel.GetTaskByID(f.OwnerID)
		utils.AssertErr(err, "", 500)
		utils.Assert(task.Publisher == userID, "permission_deny", 403)
	}
	err = s.model.RemoveFile(fileID)
	utils.AssertErr(err, "", 500)
	_, err = s.model.GetFileByHash(f.Hash)
	if err != nil {
		// 没有相同的文件
		err = libs.GetCOS().DeleteFile(f.COSName)
		utils.AssertErr(err, "", 500)
	}
}

// UpdateFileInfo 更新文件信息
func (s *fileService) UpdateFileInfo(fileID, userID primitive.ObjectID, name, description string, public bool) {
	file, err := s.model.GetFile(fileID)
	utils.AssertErr(err, "faked_file", 403)
	// 检验权限
	if file.Owner == models.FileForUser {
		utils.Assert(file.OwnerID == userID, "permission_deny")
	} else if file.Owner == models.FileForTask {
		task, err := s.taskModel.GetTaskByID(file.OwnerID)
		utils.AssertErr(err, "", 500)
		utils.Assert(task.Publisher == userID, "permission_deny")
	}
	err = s.model.SetFileInfo(fileID, name, description, public)
	utils.AssertErr(err, "", 500)
}

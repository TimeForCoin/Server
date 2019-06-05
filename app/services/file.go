package services

import (
	"github.com/TimeForCoin/Server/app/models"
)

// FileService 用户逻辑
type FileService interface {

}

// NewUserService 初始化
func newFileService() FileService {
	return &fileService{
		file: models.GetModel().File,
	}
}

type fileService struct {
	file *models.FileModel
}

func (s *fileService) AddFile(data []byte, ownerType models.OwnerType, ownID, name, description string, public bool) {

}

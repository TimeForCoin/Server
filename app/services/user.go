package services

import "github.com/TimeForCoin/Server/app/models"

// UserService 用户逻辑
type UserService interface {
	GetPong(ping string) string
}

// NewUserService 初始化
func NewUserService() UserService {
	return &userService{
		model: models.GetModel().User,
	}
}

type userService struct {
	model *models.UserModel
}

// GetPong 测试函数
func (s *userService) GetPong(ping string) string {
	if ping == "ping" {
		return "pong"
	}
	return "error"
}

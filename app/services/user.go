package services

import "github.com/TimeForCoin/Server/app/model"

// UserService 用户逻辑
type UserService interface {
	GetPong(ping string) string
}

// NewUserService 初始化
func NewUserService() UserService {
	return &userService{
		model: model.GetModel().User,
	}
}

type userService struct {
	model *model.UserModel
}

// GetPong 测试函数
func (s *userService) GetPong(ping string) string {
	if ping == "ping" {
		return "pong"
	}
	return "error"
}

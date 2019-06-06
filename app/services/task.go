package services

import (
	"errors"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/libs"
)

// TaskService 用户逻辑
type TaskService interface {
	AddTask(uid string, info models.TaskSchema) (tid string, success bool)
	SetTaskInfo(id string, info models.TaskSchema) error
	GetTaskByID(id string) (task models.TaskSchema, user models.UserSchema, attachments []models.FileSchema, err error)
}

// NewTaskService 初始化
func NewTaskService() TaskService {
	return &taskService{
		model: models.GetModel().Task,
	}
}

type taskService struct {
	model *models.TaskModel
}

func (s *taskService) AddTask(uid string, info models.TaskSchema) (tid string, success bool) {
	id, err := s.model.AddTask(uid);
	if err != nil {
		return "", false
	}
	if err = s.model.SetTaskInfoByID(id, info); err != nil {
		return id, false
	}
	return id, true
}

func (s *taskService) SetTaskInfo(id string, info models.TaskSchema) error {
	err := s.model.SetTaskInfoByID(id, info);
	libs.Assert(err == nil, "not_allow_max_finish", 403)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) GetTaskByID(id string) (task models.TaskSchema, user models.UserSchema, attachments []models.FileSchema, err error) {
	task, err = s.model.GetTaskByID(id)
	libs.Assert(err == nil, "faked_task", 403)
	if err != nil {
		return task, user, attachments, err
	}
	pid := task.Publisher
	aids := task.Attachment
	user, err = models.GetModel().User.GetUserByID(pid.Hex())
	libs.Assert(err == nil, "faked_task", 403)
	if err != nil {
		return task, user, attachments, err
	}
	for _,aid := range aids {
		attachment, err := models.GetModel().File.GetFileByID(aid.Hex())
		libs.Assert(err == nil, "faked_task", 403)
		if err != nil {
			return task, user, attachments, err
		}
		libs.Assert(attachment.Public == false && attachment.OwnerID != pid, "faked_task", 403)
		if attachment.Public == false && attachment.OwnerID != pid{
			return task, user, attachments, errors.New("invalid_session")
		}
		attachments = append(attachments, attachment)
	}
	return task, user, attachments, err
}

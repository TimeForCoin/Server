package services

import (
	"errors"
	"sort"
	"strings"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskService 用户逻辑
type TaskService interface {
	AddTask(uid string, info models.TaskSchema) (tid string, success bool)
	SetTaskInfo(id string, info models.TaskSchema) error
	GetTaskByID(id string) (task models.TaskSchema, user models.UserSchema, attachments []models.FileSchema, err error)
	GetTasks(id string, page int, size int, sortRule string, taskType string,
		status string, reward string, user string, keyword string) (totalPages int, tasks []models.TaskCard)
}

// NewUserService 初始化
func newTaskService() TaskService {
	return &taskService{
		model:     models.GetModel().Task,
		userModel: models.GetModel().User,
		oAuth:     libs.GetOAuth(),
	}
}

type taskService struct {
	model     *models.TaskModel
	userModel *models.UserModel
	oAuth     *libs.OAuthService
}

func (s *taskService) AddTask(uid string, info models.TaskSchema) (tid string, success bool) {
	id, err := s.model.AddTask(uid)
	if err != nil {
		return "", false
	}
	if err = s.model.SetTaskInfoByID(id, info); err != nil {
		return id, false
	}
	return id, true
}

func (s *taskService) SetTaskInfo(id string, info models.TaskSchema) error {
	err := s.model.SetTaskInfoByID(id, info)
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
	for _, aid := range aids {
		attachment, err := models.GetModel().File.GetFile(aid.Hex())
		libs.Assert(err == nil, "faked_task", 403)
		if err != nil {
			return task, user, attachments, err
		}
		libs.Assert(attachment.Public == false && attachment.OwnerID != pid, "faked_task", 403)
		if attachment.Public == false && attachment.OwnerID != pid {
			return task, user, attachments, errors.New("invalid_session")
		}
		attachments = append(attachments, attachment)
	}
	return task, user, attachments, err
}

// 分页获取任务列表，需要按类型/状态/酬劳类型/用户类型筛选，按关键词搜索，按不同规则排序
func (s *taskService) GetTasks(id string, page int, size int, sortRule string, taskType string,
	status string, reward string, user string, keyword string) (totalPages int, taskCards []models.TaskCard) {

	var taskTypes []models.TaskType
	split := strings.Split(taskType, ",")
	sort.Strings(split)
	if sort.SearchStrings(split, "all") != -1 || sortRule == "user" {
		taskTypes = []models.TaskType{models.TaskTypeRunning, models.TaskTypeQuestionnaire, models.TaskTypeInfo}
	} else {
		for _, str := range split {
			taskTypes = append(taskTypes, models.TaskType(str))
		}
	}

	var statuses []models.TaskStatus
	split = strings.Split(status, ",")
	sort.Strings(split)
	if sort.SearchStrings(split, "all") != -1 || sortRule == "user" {
		statuses = []models.TaskStatus{models.TaskStatusClose, models.TaskStatusDraft, models.TaskStatusFinish,
			models.TaskStatusOverdue, models.TaskStatusRun, models.TaskStatusWait}
	} else {
		for _, str := range split {
			statuses = append(statuses, models.TaskStatus(str))
		}
	}
	draft := sort.SearchStrings(split, string(models.TaskStatusDraft)) != -1

	var rewards []models.RewardType
	split = strings.Split(reward, ",")
	sort.Strings(split)
	if sort.SearchStrings(split, "all") != -1 || sortRule == "user" {
		rewards = []models.RewardType{models.RewardMoney, models.RewardPhysical, models.RewardRMB}
	} else {
		for _, str := range split {
			rewards = append(rewards, models.RewardType(str))
		}
	}

	keywords := strings.Split(keyword, " ")

	if sortRule == "new" {
		sortRule = "publish_date"
	}

	tasks, err := s.model.GetTasks(sortRule, taskTypes, statuses, rewards, keywords)
	libs.Assert(err == nil, err.Error(), 500)

	_id, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, err.Error(), 500)

	// 过滤掉非当前用户的草稿
	var tmp []models.TaskSchema
	if draft {
		for _, task := range tasks {
			if task.Status != models.TaskStatusDraft || task.Publisher == _id {
				tmp = append(tmp, task)
			}
		}
	}
	tasks = tmp

	// 筛选用户类型
	for _, task := range tasks {
		userSchema, err := s.userModel.GetUserByID(id)
		libs.Assert(err != nil, err.Error(), 500)
		if user == "certification" &&
			userSchema.Certification.Identity != models.IdentityNone || // 筛选已认证用户
			user == "credit" || // TODO 筛选高信誉用户
			user == "all" ||
			user == "" { // 筛选所有用户

			taskCard := models.TaskCard{
				ID:           task.ID.String(),
				Publisher:    task.Publisher.String(),
				Avatar:       userSchema.Info.Avatar,
				Credit:       userSchema.Data.Credit,
				Title:        task.Title,
				TopTime:      task.TopTime,
				EndDate:      task.EndDate,
				Reward:       string(task.Reward),
				RewardValue:  task.RewardValue,
				RewardObject: string(task.RewardObject),
			}

			taskCards = append(taskCards, taskCard)
		}
	}

	totalPages = len(taskCards)/size + 1

	// 选择第几页
	var upper int
	if len(tasks) < page*size {
		upper = len(tasks)
	} else {
		upper = page * size
	}
	tasks = tasks[(page-1)*size : upper]

	return
}

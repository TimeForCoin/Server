package services

import (
	"github.com/kataras/iris"
	"sort"
	"strings"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskService 用户逻辑
type TaskService interface {
	AddTask(userID primitive.ObjectID, info models.TaskSchema, publish bool)
	SetTaskInfo(taskID primitive.ObjectID, info models.TaskSchema)
	// SetTaskFile(taskID primitive.ObjectID, files []primitive.ObjectID)
	GetTaskByID(taskID primitive.ObjectID, userID string) (task TaskDetail)
	GetTasks(page, size int, sortRule, taskType,
		status , reward , user , keyword, userID string) (taskCount int, tasks []TaskDetail)
}

// NewUserService 初始化
func newTaskService() TaskService {
	return &taskService{
		model:     models.GetModel().Task,
		userModel: models.GetModel().User,
		fileModel: models.GetModel().File,
		cache:     models.GetRedis().Cache,
	}
}

type ImagesData struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type omit *struct{}
type TaskDetail struct {
	*models.TaskSchema
	// 额外项
	Publisher  models.UserBaseInfo
	Attachment []models.FileSchema
	Images     []ImagesData
	Liked      bool
	Collected  bool
	// 排除项
	LikeID omit `json:"like_id,omitempty"` // 点赞用户ID
}

type taskService struct {
	model     *models.TaskModel
	userModel *models.UserModel
	fileModel *models.FileModel
	cache     *models.CacheModel
}

func (s *taskService) AddTask(userID primitive.ObjectID, info models.TaskSchema, publish bool) {
	status := models.TaskStatusDraft
	if publish {
		status = models.TaskStatusWait
	}
	id, err := s.model.AddTask(userID, status)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	err = s.model.SetTaskInfoByID(id, info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

func (s *taskService) SetTaskInfo(taskID primitive.ObjectID, info models.TaskSchema) {
	err := s.model.SetTaskInfoByID(taskID, info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

func (s *taskService) GetTaskByID(taskID primitive.ObjectID, userID string) (task TaskDetail) {
	var err error
	taskItem, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	task.TaskSchema = &taskItem

	user, err := s.cache.GetUserBaseInfo(taskItem.Publisher)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	task.Publisher = user

	images, err := s.fileModel.GetFileByContent(taskID, models.FileImage)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	for _, i := range images {
		task.Images = append(task.Images, ImagesData{
			ID:  i.ID.Hex(),
			URL: i.URL,
		})
	}

	attachment, err := s.fileModel.GetFileByContent(taskID, models.FileFile)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	task.Attachment = attachment

	if userID != "" {
		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			task.Liked = s.cache.IsLikeTask(id, task.ID)
			task.Collected = s.cache.IsCollectTask(id, task.ID)
		}
	}

	return
}

// 分页获取任务列表，需要按类型/状态/酬劳类型/用户类型筛选，按关键词搜索，按不同规则排序
func (s *taskService) GetTasks(page, size int, sortRule , taskType ,
	status , reward , user , keyword, userID string) (totalPages int, taskCards []TaskDetail) {

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
		rewards = []models.RewardType{models.RewardMoney, models.RewardObject, models.RewardRMB}
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
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	// 过滤掉非当前用户的草稿
	var tmp []models.TaskSchema
	if draft {
		for _, task := range tasks {
			if task.Status != models.TaskStatusDraft {
				tmp = append(tmp, task)
			}
		}
	}
	tasks = tmp

	// 筛选用户类型
	//for _, task := range tasks {
	//	userSchema, err := s.userModel.GetUserByID(id)
	//	libs.Assert(err != nil, err.Error(), 500)
	//	if user == "certification" &&
	//		userSchema.Certification.Identity != models.IdentityNone || // 筛选已认证用户
	//		user == "credit" || // TODO 筛选高信誉用户
	//		user == "all" ||
	//		user == "" { // 筛选所有用户
	//
	//		taskCard := models.TaskCard{
	//			ID:           task.ID.String(),
	//			Publisher:    task.Publisher.String(),
	//			Avatar:       userSchema.Info.Avatar,
	//			Credit:       userSchema.Data.Credit,
	//			Title:        task.Title,
	//			TopTime:      task.TopTime,
	//			EndDate:      task.EndDate,
	//			Reward:       string(task.Reward),
	//			RewardValue:  task.RewardValue,
	//			RewardObject: string(task.RewardObject),
	//		}
	//
	//		taskCards = append(taskCards, taskCard)
	//	}
	//}
	for _, t := range tasks {
		var task TaskDetail
		task.TaskSchema = &t

		user, err := s.cache.GetUserBaseInfo(t.Publisher)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		task.Publisher = user

		images, err := s.fileModel.GetFileByContent(t.ID, models.FileImage)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		for _, i := range images {
			task.Images = append(task.Images, ImagesData{
				ID:  i.ID.Hex(),
				URL: i.URL,
			})
		}
		if userID != "" {
			id, err := primitive.ObjectIDFromHex(userID)
			if err != nil {
				task.Liked = s.cache.IsLikeTask(id, t.ID)
				task.Collected = s.cache.IsCollectTask(id, t.ID)
			}
		}
		taskCards = append(taskCards, task)
	}

	// TODO 待修复
	totalPages = len(taskCards)

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

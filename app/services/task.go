package services

import (
	"strings"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskService 任务服务
type TaskService interface {
	AddTask(userID primitive.ObjectID, info models.TaskSchema,
		images, attachments []primitive.ObjectID, publish bool) primitive.ObjectID
	SetTaskInfo(userID, taskID primitive.ObjectID, info models.TaskSchema,
		images, attachments []primitive.ObjectID)
	GetTaskByID(taskID primitive.ObjectID, userID string) (task TaskDetail)
	GetTasks(page, size int64, sortRule, taskType,
		status, reward, keyword, user, userID string) (taskCount int64, tasks []TaskDetail)
	RemoveTask(userID, taskID primitive.ObjectID)
	AddView(taskID primitive.ObjectID)
	ChangeLike(taskID, userID primitive.ObjectID, like bool)
	ChangeCollection(taskID, userID primitive.ObjectID, collect bool)
	ChangePlayer(taskID, userID, postUserID primitive.ObjectID, player bool) bool
	SetTaskStatusInfo(taskID, userID, postUserID primitive.ObjectID, taskStatus models.TaskStatusSchema)
	GetTaskPlayer(taskID primitive.ObjectID, status, acceptStr string, page, size int64) (taskCount int64, taskStatusList []TaskStatus)
	// 内部服务
	makeTaskDetail(task models.TaskSchema, userID string) (res TaskDetail)
}

func newTaskService() TaskService {
	return &taskService{
		model:           models.GetModel().Task,
		userModel:       models.GetModel().User,
		fileModel:       models.GetModel().File,
		cache:           models.GetRedis().Cache,
		set:             models.GetModel().Set,
		taskStatusModel: models.GetModel().TaskStatus,
	}
}

type taskService struct {
	model           *models.TaskModel
	userModel       *models.UserModel
	fileModel       *models.FileModel
	cache           *models.CacheModel
	set             *models.SetModel
	taskStatusModel *models.TaskStatusModel
}

// ImagesData 图片数据
type ImagesData struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type omit *struct{}

// TaskDetail 任务数据
type TaskDetail struct {
	*models.TaskSchema
	// 额外项
	Publisher  models.UserBaseInfo
	Attachment []models.FileSchema
	Images     []ImagesData
	Liked      bool
	Collected  bool
	Played bool
	// 排除项
	LikeID omit `json:"like_id,omitempty"` // 点赞用户ID
}

// AddTask 添加任务
func (s *taskService) AddTask(userID primitive.ObjectID, info models.TaskSchema,
	images, attachments []primitive.ObjectID, publish bool) primitive.ObjectID {
	status := models.TaskStatusDraft
	if publish {
		status = models.TaskStatusWait
	}

	taskID := primitive.NewObjectID()

	var files []FileBaseInfo
	for _, image := range images {
		files = append(files, FileBaseInfo{
			ID:   image,
			Type: models.FileImage,
		})
	}
	for _, attachment := range attachments {
		files = append(files, FileBaseInfo{
			ID:   attachment,
			Type: models.FileFile,
		})
	}
	GetServiceManger().File.BindFilesToTask(userID, taskID, files)

	id, err := s.model.AddTask(taskID, userID, status)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	err = s.model.SetTaskInfoByID(id, info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	return id
}

func (s *taskService) SetTaskInfo(userID, taskID primitive.ObjectID, info models.TaskSchema,
	images, attachments []primitive.ObjectID) {
	task, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	libs.Assert(task.Publisher == userID, "permission_deny", 403)

	libs.Assert(task.Status == models.TaskStatusDraft ||
		task.Status == models.TaskStatusWait, "not_allow_edit", 403)

	libs.Assert(string(info.Status) == "" ||
		info.Status == models.TaskStatusWait ||
		info.Status == models.TaskStatusClose ||
		info.Status == models.TaskStatusFinish, "not_allow_status", 403)
	if info.Status == models.TaskStatusWait {
		libs.Assert(task.Status == models.TaskStatusDraft, "not_allow_status", 403)
	} else if info.Status == models.TaskStatusClose || info.Status == models.TaskStatusFinish {
		libs.Assert(task.Status == models.TaskStatusWait, "not_allow_status", 403)
	}

	libs.Assert(info.MaxPlayer == 0 || info.MaxPlayer > task.PlayerCount, "not_allow_max_player", 403)

	if task.Status != models.TaskStatusDraft && task.Reward != models.RewardObject {
		libs.Assert(info.RewardValue > task.RewardValue, "not_allow_reward_value", 403)
	}

	// 更新附件
	var toRemove []primitive.ObjectID

	var imageFiles []FileBaseInfo
	if len(images) > 0 {
		oldImages, err := s.fileModel.GetFileByContent(taskID, models.FileImage)
		libs.AssertErr(err, "", 500)
		for _, image := range images {
			// 原来是否存在
			exist := false
			for _, old := range oldImages {
				if old.ID == image {
					exist = true
					break
				}
			}
			if !exist {
				imageFiles = append(imageFiles, FileBaseInfo{
					ID:   image,
					Type: models.FileImage,
				})
			}
		}
		for _, image := range oldImages {
			exist := false
			for _, file := range imageFiles {
				if file.ID == image.ID {
					exist = true
					break
				}
			}
			if !exist {
				toRemove = append(toRemove, image.ID)
			}
		}
	}

	var attachmentFiles []FileBaseInfo
	if len(attachments) > 0 {
		oldAttachment, err := s.fileModel.GetFileByContent(taskID, models.FileFile)
		libs.AssertErr(err, "", 500)
		for _, attachment := range attachments {
			// 原来是否存在
			exist := false
			for _, old := range oldAttachment {
				if old.ID == attachment {
					exist = true
					break
				}
			}
			if !exist {
				attachmentFiles = append(attachmentFiles, FileBaseInfo{
					ID:   attachment,
					Type: models.FileFile,
				})
			}
		}
		for _, attachment := range oldAttachment {
			exist := false
			for _, file := range attachmentFiles {
				if file.ID == attachment.ID {
					exist = true
					break
				}
			}
			if !exist {
				toRemove = append(toRemove, attachment.ID)
			}
		}
	}

	GetServiceManger().File.BindFilesToTask(userID, taskID, append(imageFiles, attachmentFiles...))

	err = s.model.SetTaskInfoByID(taskID, info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	// 删除无用文件
	for _, file := range toRemove {
		GetServiceManger().File.RemoveFile(file)
	}
}

// GetTaskByID 获取任务信息
func (s *taskService) GetTaskByID(taskID primitive.ObjectID, userID string) (task TaskDetail) {
	var err error
	taskItem, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	return s.makeTaskDetail(taskItem, userID)
}

// GetTasks 分页获取任务列表，需要按类型/状态/酬劳类型/用户类型筛选，按关键词搜索，按不同规则排序
func (s *taskService) GetTasks(page, size int64, sortRule, taskType,
	status, reward, keyword, user, userID string) (taskCount int64, taskCards []TaskDetail) {

	var taskTypes []models.TaskType
	var statuses []models.TaskStatus
	var rewards []models.RewardType
	var taskIDs []primitive.ObjectID
	if sortRule == "user" {
		taskTypes = []models.TaskType{models.TaskTypeRunning, models.TaskTypeQuestionnaire, models.TaskTypeInfo}
		statuses = []models.TaskStatus{models.TaskStatusClose, models.TaskStatusFinish, models.TaskStatusWait}
		rewards = []models.RewardType{models.RewardMoney, models.RewardObject, models.RewardRMB}
	} else {
		split := strings.Split(taskType, ",")
		for _, str := range split {
			if str == "all" {
				taskTypes = []models.TaskType{models.TaskTypeRunning, models.TaskTypeQuestionnaire, models.TaskTypeInfo}
				break
			}
			taskTypes = append(taskTypes, models.TaskType(str))
		}
		split = strings.Split(status, ",")
		for _, str := range split {
			if str == "all" {
				statuses = []models.TaskStatus{models.TaskStatusClose, models.TaskStatusFinish, models.TaskStatusWait}
				break
			}
			statuses = append(statuses, models.TaskStatus(str))
		}
		split = strings.Split(reward, ",")
		for _, str := range split {
			if str == "all" {
				rewards = []models.RewardType{models.RewardMoney, models.RewardObject, models.RewardRMB}
				break
			}
			rewards = append(rewards, models.RewardType(str))
		}
	}

	keywords := strings.Split(keyword, ",")
	if keyword != "" && userID != "" {
		_userID, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			//noinspection GoUnhandledErrorResult
			s.userModel.AddSearchHistory(_userID, keyword)
		}
	}

	if sortRule == "new" {
		sortRule = "publish_date"
	}

	tasks, taskCount, err := s.model.GetTasks(sortRule, taskIDs, taskTypes, statuses, rewards, keywords, user, (page-1)*size, size)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	for _, t := range tasks {
		taskCards = append(taskCards, s.makeTaskDetail(t, userID))
	}
	return
}

func (s *taskService) makeTaskDetail(task models.TaskSchema, userID string) (res TaskDetail) {
	res.TaskSchema = &task

	user, err := s.cache.GetUserBaseInfo(task.Publisher)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	res.Publisher = user

	images, err := s.fileModel.GetFileByContent(task.ID, models.FileImage)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	res.Images = []ImagesData{}
	res.Attachment = []models.FileSchema{}
	for _, i := range images {
		res.Images = append(res.Images, ImagesData{
			ID:  i.ID.Hex(),
			URL: i.URL,
		})
	}
	if userID != "" {
		id, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			res.Liked = s.cache.IsLikeTask(id, task.ID)
			res.Collected = s.cache.IsCollectTask(id, task.ID)
			_, e := s.taskStatusModel.GetTaskStatus(id, task.ID)
			res.Played = e == nil
		}
	}
	return
}

// RemoveTask 删除任务
func (s *taskService) RemoveTask(userID, taskID primitive.ObjectID) {
	task, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	libs.Assert(task.Publisher == userID, "permission_deny", 403)
	libs.Assert(task.Status == models.TaskStatusDraft, "not_allow", 403)
	err = s.model.RemoveTask(taskID)
	libs.AssertErr(err, "", 500)
	// 删除附件
	files, err := s.fileModel.GetFileByContent(taskID)
	libs.AssertErr(err, "", 500)
	for _, file := range files {
		GetServiceManger().File.RemoveFile(file.ID)
	}
}

// AddView 增加任务阅读量
func (s *taskService) AddView(taskID primitive.ObjectID) {
	err := s.model.InsertCount(taskID, models.ViewCount, 1)
	libs.AssertErr(err, "faked_task", 403)
}

// ChangeLike 改变任务点赞状态
func (s *taskService) ChangeLike(taskID, userID primitive.ObjectID, like bool) {
	task, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	libs.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	if like {
		err = s.set.AddToSet(userID, taskID, models.SetOfLikeTask)
		libs.AssertErr(err, "exist_like", 403)
		err = s.model.InsertCount(taskID, models.LikeCount, 1)
	} else {
		err = s.set.RemoveFromSet(userID, taskID, models.SetOfLikeTask)
		libs.AssertErr(err, "faked_like", 403)
		err = s.model.InsertCount(taskID, models.LikeCount, -1)
	}
	libs.AssertErr(err, "", 500)
	err = s.cache.WillUpdate(userID, models.KindOfLikeTask)
	libs.AssertErr(err, "", 500)
}

// ChangeCollection 改变收藏状态
func (s *taskService) ChangeCollection(taskID, userID primitive.ObjectID, collect bool) {
	task, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	libs.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	if collect {
		err = s.set.AddToSet(userID, taskID, models.SetOfCollectTask)
		libs.AssertErr(err, "exist_collect", 403)
		err = s.model.InsertCount(taskID, models.CollectCount, 1)
		libs.AssertErr(err, "", 500)
		err = s.userModel.AddCollectTask(userID, taskID)
	} else {
		err = s.set.RemoveFromSet(userID, taskID, models.SetOfCollectTask)
		libs.AssertErr(err, "faked_collect", 403)
		err = s.model.InsertCount(taskID, models.CollectCount, -1)
		libs.AssertErr(err, "", 500)
		err = s.userModel.RemoveCollectTask(userID, taskID)
	}
	libs.AssertErr(err, "", 500)
	err = s.cache.WillUpdate(userID, models.KindOfLikeTask)
	libs.AssertErr(err, "", 500)
}

// ChangePlayer 改变参与人员状态
func (s *taskService) ChangePlayer(taskID, userID, postUserID primitive.ObjectID, player bool) bool {
	task, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	libs.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	taskStatus, err := s.taskStatusModel.GetTaskStatus(userID, taskID)
	status := models.PlayerWait
	if player {
		libs.Assert(err != nil, "exist_player", 403)
		taskStatusID := primitive.NewObjectID()
		if task.AutoAccept {
			status = models.PlayerRunning
		}
		taskStatusID, err = s.taskStatusModel.AddTaskStatus(taskStatusID, taskID, userID, status)
		libs.AssertErr(err, "", 500)
		err = s.model.InsertCount(taskID, models.PlayerCount, 1)
		libs.AssertErr(err, "", 500)
	} else {
		libs.Assert(err == nil, "faked_player", 403)
		libs.Assert(task.Publisher == postUserID || postUserID == userID, "permission_deny", 403)
		libs.Assert(task.Status == models.TaskStatusWait, "not_allow_status", 403)
		err = s.taskStatusModel.DeleteTaskStatus(taskStatus.ID)
		libs.AssertErr(err, "", 500)
		err = s.model.InsertCount(taskID, models.PlayerCount, -1)
		libs.AssertErr(err, "", 500)
	}
	return status == models.PlayerRunning
}

func (s *taskService) SetTaskStatusInfo(taskID, userID, postUserID primitive.ObjectID, taskStatus models.TaskStatusSchema) {
	taskStatusGet, err := s.taskStatusModel.GetTaskStatus(userID, taskID)
	libs.AssertErr(err, "faked_status", 403)
	task, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	isPublisher := task.Publisher == postUserID
	libs.Assert(isPublisher || userID == postUserID, "permission_deny", 403)
	status := taskStatus.Status
	degree := taskStatus.Degree
	remark := taskStatus.Remark
	score := taskStatus.Score
	feedback := taskStatus.Feedback

	if isPublisher {
		libs.Assert(score == 0 && feedback == "" &&
			(status == "" || status == models.PlayerRunning || status == models.PlayerFailure || status == models.PlayerFinish), "permission_deny", 403)
		if taskStatus.Status != models.PlayerFinish && taskStatus.Status != models.PlayerClose && taskStatus.Status != models.PlayerFailure {
			libs.Assert(degree == 0 && remark == "", "not_allow_status", 403)
		}
	} else {
		libs.Assert(degree == 0 && remark == "" && (status == "" || status == models.PlayerGiveUp), "permission_deny", 403)
		if taskStatus.Status != models.PlayerFinish && taskStatus.Status != models.PlayerClose && taskStatus.Status != models.PlayerFailure {
			libs.Assert(score == 0 && feedback == "", "not_allow_status", 403)
		}
	}
	err = s.taskStatusModel.SetTaskStatus(taskStatusGet.ID, taskStatus)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

func (s *taskService) GetTaskPlayer(taskID primitive.ObjectID, status, acceptStr string, page, size int64) (taskCount int64, taskStatusList []TaskStatus) {
	_, err := s.model.GetTaskByID(taskID)
	libs.AssertErr(err, "faked_task", 403)
	var statuses []models.PlayerStatus
	split := strings.Split(status, ",")
	for _, str := range split {
		if str == "all" {
			statuses = []models.PlayerStatus{models.PlayerWait, models.PlayerRefuse, models.PlayerClose, models.PlayerRunning, models.PlayerFinish, models.PlayerGiveUp, models.PlayerFailure}
			break
		}
		statuses = append(statuses, models.PlayerStatus(str))
	}
	if acceptStr == "true" {
		for i := 0; i < len(statuses); {
			if statuses[i] == models.PlayerWait || statuses[i] == models.PlayerRefuse {
				statuses = append(statuses[:i], statuses[i+1:]...)
			} else {
				i++
			}
		}
	} else if acceptStr == "false" {
		for i := 0; i < len(statuses); {
			if statuses[i] != models.PlayerWait && statuses[i] != models.PlayerRefuse {
				statuses = append(statuses[:i], statuses[i+1:]...)
			} else {
				i++
			}
		}
	}

	taskStatuses, taskCount, err := s.taskStatusModel.GetTaskStatusListByTaskID(taskID, statuses, (page-1)*size, size)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	for i, t := range taskStatuses {
		var taskStatus TaskStatus
		taskStatus.TaskStatusSchema = &taskStatuses[i]

		userPlayer, err := s.cache.GetUserBaseInfo(t.Player)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		taskStatus.Player = userPlayer

		taskStatusList = append(taskStatusList, taskStatus)
	}
	return
}

package services

import (
	"strings"
	"time"

	"github.com/TimeForCoin/Server/app/utils"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskService 任务服务
type TaskService interface {
	AddTask(userID primitive.ObjectID, info models.TaskSchema,
		images, attachments []primitive.ObjectID, publish bool) primitive.ObjectID
	SetTaskInfo(userID, taskID primitive.ObjectID, info models.TaskSchema,
		images, attachments []primitive.ObjectID)
	GetTaskByID(taskID primitive.ObjectID, userID string, biref bool) (task TaskDetail)
	GetTasks(page, size int64, sortRule, taskType,
		status, reward, keyword, user, userID string, biref bool) (taskCount int64, tasks []TaskDetail)
	RemoveTask(userID, taskID primitive.ObjectID)
	AddView(taskID primitive.ObjectID)
	ChangeLike(taskID, userID primitive.ObjectID, like bool)
	ChangeCollection(taskID, userID primitive.ObjectID, collect bool)
	AddPlayer(taskID, userID primitive.ObjectID, note string) bool
	GetTaskStatus(taskID, userID, postUserID primitive.ObjectID) (taskStatusList TaskStatus)
	SetTaskStatusInfo(taskID, userID, postUserID primitive.ObjectID, taskStatus models.TaskStatusSchema)
	GetTaskPlayer(taskID primitive.ObjectID, status string, page, size int64) (taskCount int64, taskStatusList []TaskStatus)
	GetQRCode(taskID primitive.ObjectID) string
	// 内部服务
	makeTaskDetail(task models.TaskSchema, userID string, biref bool) (res TaskDetail)
}

func newTaskService() TaskService {
	return &taskService{
		model:           models.GetModel().Task,
		userModel:       models.GetModel().User,
		fileModel:       models.GetModel().File,
		cache:           models.GetRedis().Cache,
		setModel:        models.GetModel().Set,
		taskStatusModel: models.GetModel().TaskStatus,
		messageModel:    models.GetModel().Message,
		logModel:        models.GetModel().Log,
	}
}

type taskService struct {
	model           *models.TaskModel
	userModel       *models.UserModel
	fileModel       *models.FileModel
	cache           *models.CacheModel
	setModel        *models.SetModel
	taskStatusModel *models.TaskStatusModel
	messageModel    *models.MessageModel
	logModel        *models.LogModel
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
	Played     bool
	// 排除项
	LikeID omit `json:"like_id,omitempty"` // 点赞用户ID
}

// AddTask 添加任务
func (s *taskService) AddTask(userID primitive.ObjectID, info models.TaskSchema,
	images, attachments []primitive.ObjectID, publish bool) primitive.ObjectID {
	status := models.TaskStatusDraft
	user, err := s.userModel.GetUserByID(userID)
	utils.AssertErr(err, "", 500)
	if publish {
		status = models.TaskStatusWait
		utils.Assert(float32(user.Data.Money) > info.RewardValue*float32(info.MaxPlayer)+1, "no_money", 403)
	} else {
		utils.Assert(float32(user.Data.Money) > 1, "no_money", 403)
	}
	utils.Assert(float32(user.Data.Value) > 2, "no_value", 403)

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
	utils.AssertErr(err, "", iris.StatusInternalServerError)

	err = s.model.SetTaskInfoByID(id, info)
	utils.AssertErr(err, "", iris.StatusInternalServerError)

	if publish {
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Money: -int64(info.RewardValue)*info.MaxPlayer - 1,
			Value: -2,
		})
		utils.AssertErr(err, "", iris.StatusInternalServerError)
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeMoney)
		err = s.logModel.SetValue(logID, -int64(info.RewardValue)*info.MaxPlayer-1)
		err = s.logModel.SetMsg(logID, "funish task")
		logID, err = s.logModel.AddLog(userID, taskID, models.LogTypeValue)
		err = s.logModel.SetValue(logID, -2)
		err = s.logModel.SetMsg(logID, "funish task")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	} else {
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Money: -1,
			Value: -2,
		})
		utils.AssertErr(err, "", iris.StatusInternalServerError)
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeMoney)
		err = s.logModel.SetValue(logID, -1)
		err = s.logModel.SetMsg(logID, "funish task")
		logID, err = s.logModel.AddLog(userID, taskID, models.LogTypeValue)
		err = s.logModel.SetValue(logID, -2)
		err = s.logModel.SetMsg(logID, "funish task")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	}

	return id
}

func (s *taskService) SetTaskInfo(userID, taskID primitive.ObjectID, info models.TaskSchema,
	images, attachments []primitive.ObjectID) {
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	utils.Assert(task.Publisher == userID, "permission_deny", 403)

	// 状态修改
	if info.Status == models.TaskStatusClose {
		// 关闭任务
		utils.Assert(task.Status == models.TaskStatusWait, "not_allow_status", 403)

		taskStatus, _, err := s.taskStatusModel.GetTaskStatusListByTaskID(taskID, []models.PlayerStatus{}, 0, 0)
		utils.AssertErr(err, "", 500)
		// 发送通知消息
		for _, status := range taskStatus {
			_, err = s.messageModel.AddMessage(status.Player, models.MessageTypeTask, models.MessageSchema{
				UserID: taskID,
				Title:  "任务已关闭",
			})
			utils.AssertErr(err, "", 500)
			err = s.taskStatusModel.SetTaskStatus(status.ID, models.TaskStatusSchema{
				Status: models.PlayerClose,
			})
			utils.AssertErr(err, "", 500)
		}

		err = s.model.SetTaskInfoByID(taskID, models.TaskSchema{
			Status: models.TaskStatusClose,
		})
		utils.AssertErr(err, "", iris.StatusInternalServerError)
		return
	} else if info.Status == models.TaskStatusWait {
		// 发布任务
		utils.Assert(task.Status == models.TaskStatusDraft, "not_allow_status", 403)
		info.PublishDate = time.Now().Unix()
		user, err := s.userModel.GetUserByID(userID)
		utils.AssertErr(err, "", 500)
		utils.Assert(float32(user.Data.Money) > info.RewardValue*float32(info.MaxPlayer), "no_money", 403)
		utils.Assert(float32(user.Data.Value) > 2, "no_value", 403)
	} else if info.Status == models.TaskStatusFinish {
		// 任务已完成
		players, _, err := s.taskStatusModel.GetTaskStatusListByTaskID(taskID, []models.PlayerStatus{}, 0, 0)
		utils.AssertErr(err, "", 500)
		for _, status := range players {
			utils.Assert(status.Status != models.PlayerRunning && status.Status != models.PlayerWait, "not_allow_finish", 403)
		}
	} else if info.Status != models.TaskStatusDraft && info.Status != "" {
		utils.Assert(false, "not_allow_status", 403)
	}

	utils.Assert(info.MaxPlayer == 0 || info.MaxPlayer > task.PlayerCount, "not_allow_max_player", 403)

	if info.Type != "" {
		if task.Type == models.TaskTypeQuestionnaire || info.Type == models.TaskTypeQuestionnaire {
			utils.Assert(info.Type == task.Type, "not_allow_change_type")
		}
	}

	addMoney := 0
	if task.Status != models.TaskStatusDraft {
		utils.Assert(info.Reward == "" || task.Reward == info.Reward, "not_allow_change_reward_type", 403)
		if task.Reward != models.RewardObject {
			if info.RewardValue != 0 {
				utils.Assert(info.RewardValue >= task.RewardValue, "not_allow_reward_value", 403)
				addMoney = int(info.RewardValue - task.RewardValue)
			}
		}
	}

	// 更新附件
	var toRemove []primitive.ObjectID

	var imageFiles []FileBaseInfo
	if len(images) > 0 {
		oldImages, err := s.fileModel.GetFileByContent(taskID, models.FileImage)
		utils.AssertErr(err, "", 500)
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
			for _, file := range images {
				if file == image.ID {
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
		utils.AssertErr(err, "", 500)
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
			for _, file := range attachments {
				if file == attachment.ID {
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
	utils.AssertErr(err, "", iris.StatusInternalServerError)

	if info.Status == models.TaskStatusWait {
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Money: -int64(info.RewardValue) * info.MaxPlayer,
		})
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeMoney)
		err = s.logModel.SetValue(logID, -int64(info.RewardValue)*info.MaxPlayer)
		err = s.logModel.SetMsg(logID, "set task info")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	} else if addMoney != 0 {
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Money: -int64(addMoney),
		})
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeMoney)
		err = s.logModel.SetValue(logID, -int64(addMoney))
		err = s.logModel.SetMsg(logID, "set task info")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	}

	// 删除无用文件
	for _, file := range toRemove {
		GetServiceManger().File.RemoveFile(file)
	}
}

// GetTaskByID 获取任务信息
func (s *taskService) GetTaskByID(taskID primitive.ObjectID, userID string, biref bool) (task TaskDetail) {
	var err error
	taskItem, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	return s.makeTaskDetail(taskItem, userID, biref)
}

// GetTasks 分页获取任务列表，需要按类型/状态/酬劳类型/用户类型筛选，按关键词搜索，按不同规则排序
func (s *taskService) GetTasks(page, size int64, sortRule, taskType,
	status, reward, keyword, user, userID string, biref bool) (taskCount int64, taskCards []TaskDetail) {

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
	utils.AssertErr(err, "", iris.StatusInternalServerError)

	for _, t := range tasks {
		taskCards = append(taskCards, s.makeTaskDetail(t, userID, biref))
	}
	return
}

func (s *taskService) makeTaskDetail(task models.TaskSchema, userID string, biref bool) (res TaskDetail) {
	res.TaskSchema = &task

	user, err := s.cache.GetUserBaseInfo(task.Publisher)
	utils.AssertErr(err, "", iris.StatusInternalServerError)
	res.Publisher = user

	images, err := s.fileModel.GetFileByContent(task.ID, models.FileImage)
	utils.AssertErr(err, "", iris.StatusInternalServerError)
	res.Images = []ImagesData{}
	for _, i := range images {
		res.Images = append(res.Images, ImagesData{
			ID:  i.ID.Hex(),
			URL: i.URL,
		})
		if biref {
			break
		}
	}
	if !biref {
		attachments, err := s.fileModel.GetFileByContent(task.ID, models.FileFile)
		utils.AssertErr(err, "", iris.StatusInternalServerError)
		res.Attachment = []models.FileSchema{}
		for _, attachment := range attachments {
			res.Attachment = append(res.Attachment, attachment)
		}
	}
	if userID != "" {
		id, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			res.Liked = s.cache.IsLikeTask(id, task.ID)
			res.Collected = s.cache.IsCollectTask(id, task.ID)
			status, e := s.taskStatusModel.GetTaskStatus(id, task.ID)
			res.Played = e == nil && status.Status != models.PlayerGiveUp
		}
	}
	return
}

// RemoveTask 删除任务
func (s *taskService) RemoveTask(userID, taskID primitive.ObjectID) {
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	utils.Assert(task.Publisher == userID, "permission_deny", 403)
	utils.Assert(task.Status == models.TaskStatusDraft, "not_allow", 403)
	err = s.model.RemoveTask(taskID)
	utils.AssertErr(err, "", 500)
	// 删除附件
	files, err := s.fileModel.GetFileByContent(taskID)
	utils.AssertErr(err, "", 500)
	for _, file := range files {
		GetServiceManger().File.RemoveFile(file.ID)
	}
}

// AddView 增加任务阅读量
func (s *taskService) AddView(taskID primitive.ObjectID) {
	err := s.model.InsertCount(taskID, models.ViewCount, 1)
	utils.AssertErr(err, "faked_task", 403)
}

// ChangeLike 改变任务点赞状态
func (s *taskService) ChangeLike(taskID, userID primitive.ObjectID, like bool) {
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	utils.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	if like {
		err = s.setModel.AddToSet(userID, taskID, models.SetOfLikeTask)
		utils.AssertErr(err, "exist_like", 403)
		err = s.model.InsertCount(taskID, models.LikeCount, 1)
	} else {
		err = s.setModel.RemoveFromSet(userID, taskID, models.SetOfLikeTask)
		utils.AssertErr(err, "faked_like", 403)
		err = s.model.InsertCount(taskID, models.LikeCount, -1)
	}
	utils.AssertErr(err, "", 500)
	err = s.cache.WillUpdate(userID, models.KindOfLikeTask)
	utils.AssertErr(err, "", 500)
}

// ChangeCollection 改变收藏状态
func (s *taskService) ChangeCollection(taskID, userID primitive.ObjectID, collect bool) {
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	utils.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	if collect {
		err = s.setModel.AddToSet(userID, taskID, models.SetOfCollectTask)
		utils.AssertErr(err, "exist_collect", 403)
		err = s.model.InsertCount(taskID, models.CollectCount, 1)
	} else {
		err = s.setModel.RemoveFromSet(userID, taskID, models.SetOfCollectTask)
		utils.AssertErr(err, "faked_collect", 403)
		err = s.model.InsertCount(taskID, models.CollectCount, -1)
	}
	utils.AssertErr(err, "", 500)
	err = s.cache.WillUpdate(userID, models.KindOfCollectTask)
	utils.AssertErr(err, "", 500)
}

// AddPlayer 增加参与人员
func (s *taskService) AddPlayer(taskID, userID primitive.ObjectID, note string) bool {
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	utils.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	utils.Assert(task.PlayerCount < task.MaxPlayer, "max_player", 403)
	user, err := s.userModel.GetUserByID(userID)
	utils.AssertErr(err, "", 500)
	utils.Assert(user.Data.Value > 1, "no_value", 403)
	taskStatus, err := s.taskStatusModel.GetTaskStatus(userID, taskID)

	status := models.PlayerWait
	if task.AutoAccept {
		status = models.PlayerRunning
	}
	if err != nil {
		// 不存在记录
		err = s.taskStatusModel.AddTaskStatus(taskID, userID, status, note)
	} else {
		// 存在记录
		utils.Assert(taskStatus.Status == models.PlayerGiveUp, "not_allow_status", 403)
		err = s.taskStatusModel.SetTaskStatus(taskStatus.ID, models.TaskStatusSchema{
			Status: status,
			Note:   note,
		})
	}
	utils.AssertErr(err, "", 500)
	err = s.model.InsertCount(taskID, models.PlayerCount, 1)
	utils.AssertErr(err, "", 500)

	msg := user.Info.Nickname
	if status == models.PlayerRunning {
		msg += "申请加入任务"
	} else {
		msg += "加入任务"
	}

	_, err = s.messageModel.AddMessage(task.Publisher, models.MessageTypeTask, models.MessageSchema{
		UserID:  taskID,
		Title:   msg,
		Content: taskStatus.Note,
		About:   userID,
	})
	utils.AssertErr(err, "", 500)

	err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
		Value: -1,
	})
	logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeValue)
	err = s.logModel.SetValue(logID, -1)
	err = s.logModel.SetMsg(logID, "Add Player")
	utils.AssertErr(err, "", iris.StatusInternalServerError)
	return status == models.PlayerRunning
}

// SetTaskStatusInfo 设置参与任务信息
func (s *taskService) SetTaskStatusInfo(taskID, userID, postUserID primitive.ObjectID, taskStatus models.TaskStatusSchema) {
	taskStatusGet, err := s.taskStatusModel.GetTaskStatus(userID, taskID)
	utils.AssertErr(err, "faked_status", 403)
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	isPublisher := task.Publisher == postUserID
	utils.Assert(isPublisher || userID == postUserID, "permission_deny", 403)

	// 状态修改
	if taskStatus.Status == models.PlayerRunning {
		utils.Assert(isPublisher, "permission_deny", 403)
		utils.Assert(taskStatusGet.Status == models.PlayerWait, "not_allow_status", 403)
	} else if taskStatus.Status == models.PlayerFinish || taskStatus.Status == models.PlayerFailure {
		utils.Assert(isPublisher, "permission_deny", 403)
		utils.Assert(taskStatusGet.Status == models.PlayerRunning, "not_allow_status", 403)
	} else if taskStatus.Status == models.PlayerRefuse {
		utils.Assert(isPublisher, "permission_deny", 403)
		utils.Assert(taskStatusGet.Status == models.PlayerWait, "not_allow_status", 403)
	} else if taskStatus.Status == models.PlayerGiveUp {
		utils.Assert(taskStatusGet.Status == models.PlayerRunning || taskStatusGet.Status == models.PlayerWait, "not_allow_status", 403)
	} else if string(taskStatus.Status) != "" {
		utils.Assert(false, "not_allow_status", 403)
	}

	// 允许修改的信息
	if isPublisher {
		utils.Assert(taskStatus.Score == 0 && taskStatus.Feedback == "", "permission_deny", 403)
	} else {
		utils.Assert(taskStatus.Degree == 0 && taskStatus.Remark == "", "permission_deny", 403)
	}

	err = s.taskStatusModel.SetTaskStatus(taskStatusGet.ID, models.TaskStatusSchema{
		Status:   taskStatus.Status,
		Degree:   taskStatus.Degree,
		Remark:   taskStatus.Remark,
		Score:    taskStatus.Score,
		Feedback: taskStatus.Feedback,
	})
	utils.AssertErr(err, "", iris.StatusInternalServerError)

	// 发送消息
	if taskStatus.Status == models.PlayerRefuse {
		_, err = s.messageModel.AddMessage(userID, models.MessageTypeTask, models.MessageSchema{
			UserID:  taskID,
			Title:   "你的任务申请被拒绝了",
			Content: taskStatus.Note,
		})
		utils.AssertErr(err, "", 500)
	} else if taskStatus.Status == models.PlayerRunning {
		_, err = s.messageModel.AddMessage(userID, models.MessageTypeTask, models.MessageSchema{
			UserID:  taskID,
			Title:   "你的任务申请已通过",
			Content: taskStatus.Note,
		})
		utils.AssertErr(err, "", 500)
	} else if taskStatus.Status == models.PlayerFinish {
		_, err = s.messageModel.AddMessage(userID, models.MessageTypeTask, models.MessageSchema{
			UserID:  taskID,
			Title:   "恭喜你，任务已完成",
			Content: taskStatus.Note,
		})
		utils.AssertErr(err, "", 500)
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Money: int64(task.RewardValue),
			Value: 5,
		})
		utils.AssertErr(err, "", 500)
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeMoney)
		err = s.logModel.SetValue(logID, int64(task.RewardValue))
		err = s.logModel.SetMsg(logID, "funish task")
		logID, err = s.logModel.AddLog(userID, taskID, models.LogTypeValue)
		err = s.logModel.SetValue(logID, 5)
		err = s.logModel.SetMsg(logID, "funish task")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	} else if taskStatus.Status == models.PlayerFinish {
		_, err = s.messageModel.AddMessage(userID, models.MessageTypeTask, models.MessageSchema{
			UserID:  taskID,
			Title:   "很遗憾，任务已失败",
			Content: taskStatus.Note,
		})
		utils.AssertErr(err, "", 500)
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Value: -1,
		})
		utils.AssertErr(err, "", 500)
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeValue)
		err = s.logModel.SetValue(logID, -1)
		err = s.logModel.SetMsg(logID, "Player Finish")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	} else if taskStatus.Status == models.PlayerGiveUp {
		user := GetServiceManger().User.GetUserBaseInfo(userID)
		_, err = s.messageModel.AddMessage(task.Publisher, models.MessageTypeTask, models.MessageSchema{
			UserID:  taskID,
			Title:   user.Nickname + "放弃任务",
			Content: taskStatus.Note,
			About:   userID,
		})
		utils.AssertErr(err, "", 500)
		err = s.model.InsertCount(taskID, models.PlayerCount, -1)
		utils.AssertErr(err, "", 500)
		err = s.userModel.UpdateUserDataCount(userID, models.UserDataCount{
			Value: -3,
		})
		utils.AssertErr(err, "", 500)
		logID, err := s.logModel.AddLog(userID, taskID, models.LogTypeValue)
		err = s.logModel.SetValue(logID, -3)
		err = s.logModel.SetMsg(logID, "Player Give Up")
		utils.AssertErr(err, "", iris.StatusInternalServerError)
	}
}

// SetTaskStatus 设置任务状态
func (s *taskService) GetTaskStatus(taskID, userID, postUserID primitive.ObjectID) (taskStatus TaskStatus) {
	taskStatusGet, err := s.taskStatusModel.GetTaskStatus(userID, taskID)
	utils.AssertErr(err, "faked_status", 403)
	task, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	utils.Assert(task.Publisher == postUserID || userID == postUserID, "permission_deny", 403)
	taskStatus.TaskStatusSchema = &taskStatusGet
	userPlayer, err := s.cache.GetUserBaseInfo(taskStatusGet.Player)
	utils.AssertErr(err, "", iris.StatusInternalServerError)
	taskStatus.Player = userPlayer
	return
}

func (s *taskService) GetTaskPlayer(taskID primitive.ObjectID, status string, page, size int64) (taskCount int64, taskStatusList []TaskStatus) {
	_, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	var statuses []models.PlayerStatus
	split := strings.Split(status, ",")
	for _, str := range split {
		if str == "all" {
			statuses = []models.PlayerStatus{models.PlayerWait, models.PlayerRefuse, models.PlayerClose, models.PlayerRunning, models.PlayerFinish, models.PlayerGiveUp, models.PlayerFailure}
			break
		}
		statuses = append(statuses, models.PlayerStatus(str))
	}

	taskStatuses, taskCount, err := s.taskStatusModel.GetTaskStatusListByTaskID(taskID, statuses, (page-1)*size, size)
	utils.AssertErr(err, "", iris.StatusInternalServerError)
	for i, t := range taskStatuses {
		var taskStatus TaskStatus
		taskStatus.TaskStatusSchema = &taskStatuses[i]

		userPlayer, err := s.cache.GetUserBaseInfo(t.Player)
		utils.AssertErr(err, "", iris.StatusInternalServerError)
		taskStatus.Player = userPlayer

		taskStatusList = append(taskStatusList, taskStatus)
	}
	return
}

// 获取小程序码
func (s *taskService) GetQRCode(taskID primitive.ObjectID) string {
	_, err := s.model.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_task", 403)
	image, err := s.fileModel.GetFileByContent(taskID, models.FileQRCode)
	utils.AssertErr(err, "", 500)
	if len(image) == 0 {
		// 不存在记录
		base64, err := libs.GetWeChat().MakeImage(taskID.Hex())
		utils.AssertErr(err, "", 500)

		fileID := primitive.NewObjectID()
		// 上传到腾讯云
		cosName := "qr_code-" + taskID.Hex() + "-" + fileID.Hex() + ".png"
		url, err := libs.GetCOS().SaveBase64File(cosName, base64)
		utils.AssertErr(err, "", 500)

		// 保存到数据库
		err = s.fileModel.AddFile(models.FileSchema{
			ID:      fileID,
			OwnerID: taskID,
			Owner:   models.FileForTask,
			Type:    models.FileQRCode,
			URL:     url,
			Size:    int64(len(base64)),
			Public:  true,
			Used:    1,
			Hash:    fileID.Hex(),
			COSName: cosName,
		})
		utils.AssertErr(err, "", 500)
		return url
	}
	return image[0].URL
}

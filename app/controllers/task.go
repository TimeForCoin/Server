package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskController 任务相关API
type TaskController struct {
	BaseController
	Service services.TaskService
}

// BindTaskController 绑定任务控制器
func BindTaskController(app *iris.Application) {
	taskService := services.GetServiceManger().Task

	taskRoute := mvc.New(app.Party("/tasks"))
	taskRoute.Register(taskService, getSession().Start)
	taskRoute.Handle(new(TaskController))
}

// AddTaskReq 添加任务请求
type AddTaskReq struct {
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Images       []string `json:"images"`
	Attachment   []string `json:"attachment"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	Reward       string   `json:"reward"`
	RewardValue  float32  `json:"reward_value"`
	RewardObject string   `json:"reward_object"`
	Location     []string `json:"location"`
	Tags         []string `json:"tags"`
	StartDate    int64    `json:"start_date"`
	EndDate      int64    `json:"end_date"`
	MaxPlayer    int64    `json:"max_player"`
	AutoAccept   bool     `json:"auto_accept"`
	Publish      bool     `json:"publish"`
}

func validTask(req AddTaskReq, new bool) {
	if req.Type != "" || new {
		libs.Assert(models.TaskType(req.Type) == models.TaskTypeInfo ||
			models.TaskType(req.Type) == models.TaskTypeQuestionnaire ||
			models.TaskType(req.Type) == models.TaskTypeRunning, "invalid_type", 400)
	}

	if new {
		libs.Assert(req.Title != "", "invalid_title", 400)
		libs.Assert(req.Content != "", "invalid_content", 400)
		libs.CheckReward(req.Reward, req.RewardObject, req.RewardValue)
		libs.Assert(req.MaxPlayer > 0, "invalid_max_player", 400)
		libs.CheckDateDuring(req.StartDate, req.EndDate)
	} else {
		libs.Assert(req.MaxPlayer >= 0, "invalid_max_player", 400)
		if req.StartDate != 0 || req.EndDate != 0 {
			libs.CheckDateDuring(req.StartDate, req.EndDate)
		}
	}



	libs.Assert(len(req.Title) < 64, "title_too_long", 403)
	libs.Assert(len(req.Content) < 512, "content_too_long", 403)
	libs.Assert(len(req.RewardObject) < 32, "reward_object_too_long", 403)

	for _, l := range req.Location {
		libs.Assert(len(l) < 64, "location_too_long", 403)
	}

	for _, t := range req.Tags {
		libs.Assert(len(t) < 16, "tag_too_long", 403)
	}

	for _, file := range req.Images {
		_, err := primitive.ObjectIDFromHex(file)
		libs.AssertErr(err, "invalid_file", 400)
	}

	for _, attachment := range req.Attachment {
		_, err := primitive.ObjectIDFromHex(attachment)
		libs.AssertErr(err, "invalid_file", 400)
	}
}

// Post 添加任务
func (c *TaskController) Post() int {
	id := c.checkLogin()
	req := AddTaskReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	validTask(req, true)

	taskType := models.TaskType(req.Type)
	taskReward := models.RewardType(req.Reward)

	var images []primitive.ObjectID
	for _, file := range req.Images {
		fileID, _ := primitive.ObjectIDFromHex(file)
		images = append(images, fileID)
	}
	var attachments []primitive.ObjectID
	for _, attachment := range req.Attachment {
		fileID, _ := primitive.ObjectIDFromHex(attachment)
		attachments = append(attachments, fileID)
	}

	taskInfo := models.TaskSchema{
		Title:        req.Title,
		Type:         taskType,
		Content:      req.Content,
		Location:     req.Location,
		Tags:         req.Tags,
		Reward:       taskReward,
		RewardValue:  req.RewardValue,
		RewardObject: req.RewardObject,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		MaxPlayer:    req.MaxPlayer,
		AutoAccept:   req.AutoAccept,
	}
	taskID := c.Service.AddTask(id, taskInfo, images, attachments, req.Publish)
	c.JSON(struct {
		ID string `json:"id"`
	}{
		ID: taskID.Hex(),
	})
	return iris.StatusOK
}

// GetBy 获取指定任务详情
func (c *TaskController) GetBy(id string) int {
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	briefParam := c.Ctx.URLParamDefault("brief", "false")
	biref := false
	if briefParam == "true" {
		biref = true
	}

	task := c.Service.GetTaskByID(_id, c.Session.GetString("id"), biref)
	c.Service.AddView(_id)
	c.JSON(task)
	return iris.StatusOK
}

// PutBy 修改指定任务信息
func (c *TaskController) PutBy(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	req := AddTaskReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	validTask(req, false)

	var images []primitive.ObjectID
	for _, file := range req.Images {
		fileID, _ := primitive.ObjectIDFromHex(file)
		images = append(images, fileID)
	}
	var attachments []primitive.ObjectID
	for _, attachment := range req.Attachment {
		fileID, _ := primitive.ObjectIDFromHex(attachment)
		attachments = append(attachments, fileID)
	}

	libs.Assert(req.Status == "" || libs.IsTaskStatus(req.Status), "invalid_status", 400)

	taskInfo := models.TaskSchema{
		Title:        req.Title,
		Content:      req.Content,
		Location:     req.Location,
		Tags:         req.Tags,
		Status:       models.TaskStatus(req.Status),
		RewardValue:  req.RewardValue,
		RewardObject: req.RewardObject,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		MaxPlayer:    req.MaxPlayer,
		AutoAccept:   req.AutoAccept,
	}
	c.Service.SetTaskInfo(userID, taskID, taskInfo, images, attachments)
	return iris.StatusOK
}

// GetTasksReq 获取任务列表请求
type GetTasksReq struct {
	Page    int
	Size    int
	Sort    string
	Type    string
	Status  string
	Reward  string
	User    string
	Keyword string
}

// TasksListRes 任务列表数据
type TasksListRes struct {
	Pagination PaginationRes
	Tasks      []services.TaskDetail
}

// Get 获取任务列表
func (c *TaskController) Get() int {
	page, size := c.getPaginationData()

	sort := c.Ctx.URLParamDefault("sort", "new")
	taskType := c.Ctx.URLParamDefault("type", "all")
	status := c.Ctx.URLParamDefault("status", "wait")
	reward := c.Ctx.URLParamDefault("reward", "all")
	keyword := c.Ctx.URLParamDefault("keyword", "")
	user := c.Ctx.URLParamDefault("user", "")
	birefParam := c.Ctx.URLParamDefault("biref", "false")
	biref := false
	if birefParam == "true" {
		biref = true
	}

	if status == string(models.TaskStatusDraft) {
		libs.Assert(user == "me" || user == "", "not_allow_other_draft", 403)
		user = "me"
	}

	if user == "me" {
		// 自己的任务
		id := c.checkLogin()
		user = id.Hex()

	} else if user != "" {
		// 筛选用户
		_, err := primitive.ObjectIDFromHex(user)
		libs.AssertErr(err, "invalid_user", 403)
	}

	taskCount, tasksData := c.Service.GetTasks(page, size, sort,
		taskType, status, reward, keyword, user, c.Session.GetString("id"), biref)

	if tasksData == nil {
		tasksData = []services.TaskDetail{}
	}

	res := TasksListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: taskCount,
		},
		Tasks: tasksData,
	}
	c.JSON(res)
	return iris.StatusOK
}

// DeleteBy 删除草稿任务
func (c *TaskController) DeleteBy(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.RemoveTask(userID, taskID)
	return iris.StatusOK
}

// PostByLike 点赞任务
func (c *TaskController) PostByLike(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.ChangeLike(taskID, userID, true)
	return iris.StatusOK
}

// DeleteByLike 取消点赞任务
func (c *TaskController) DeleteByLike(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.ChangeLike(taskID, userID, false)
	return iris.StatusOK
}

// PostByCollect 添加收藏
func (c *TaskController) PostByCollect(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.ChangeCollection(taskID, userID, true)
	return iris.StatusOK
}

// DeleteByCollect 删除收藏
func (c *TaskController) DeleteByCollect(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.ChangeCollection(taskID, userID, false)
	return iris.StatusOK
}

// PostPlayerReq 增加人员请求
type PostPlayerReq struct {
	Note string
}

// PostByPlayer 增加任务参与人员
func (c *TaskController) PostByPlayer(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	req := PostPlayerReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	accept := c.Service.AddPlayer(taskID, userID, req.Note)
	res := "wait"
	if accept {
		res = "accept"
	}
	c.JSON(struct {
		Result string
	}{
		Result: res,
	})
	return iris.StatusOK
}

// TaskStatusReq 任务参与状态请求
type TaskStatusReq struct {
	Status   string `json:"status"`
	Note     string `json:"note"`
	Degree   int    `json:"degree"`
	Remark   string `json:"remark"`
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// GetByPlayerBy 获取用户任务状态
func (c *TaskController) GetByPlayerBy(id, userIDString string) int {
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	var userID primitive.ObjectID
	if userIDString == "me" {
		userID = c.checkLogin()
	} else {
		userID, err = primitive.ObjectIDFromHex(userIDString)
		libs.AssertErr(err, "invalid_id", 400)
	}

	status := c.Service.GetTaskStatus(taskID, userID, c.checkLogin())

	c.JSON(struct {
		Data services.TaskStatus
	}{
		Data: status,
	})

	return iris.StatusOK
}

// PutByPlayerBy 修改任务参与者状态
func (c *TaskController) PutByPlayerBy(id, userIDString string) int {
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	var userID primitive.ObjectID
	if userIDString == "me" {
		userID = c.checkLogin()
	} else {
		userID, err = primitive.ObjectIDFromHex(userIDString)
		libs.AssertErr(err, "invalid_id", 400)
	}

	postUserID := c.checkLogin()

	req := TaskStatusReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	var status models.PlayerStatus
	if req.Status != "" {
		status = models.PlayerStatus(req.Status)
		libs.Assert(status == models.PlayerWait || status == models.PlayerRefuse ||
			status == models.PlayerClose || status == models.PlayerRunning ||
			status == models.PlayerFinish || status == models.PlayerGiveUp ||
			status == models.PlayerFailure, "invalid_status", 403)
	}

	taskStatusInfo := models.TaskStatusSchema{
		Status:   status,
		Note:     req.Note,
		Degree:   req.Degree,
		Remark:   req.Remark,
		Score:    req.Score,
		Feedback: req.Feedback,
	}

	c.Service.SetTaskStatusInfo(taskID, userID, postUserID, taskStatusInfo)
	return iris.StatusOK
}

// PlayerListRes 任务参与用户数据
type PlayerListRes struct {
	Pagination PaginationRes
	Data       []services.TaskStatus
}

// GetByPlayer 获取任务参与者
func (c *TaskController) GetByPlayer(id string) int {
	libs.Assert(id != "", "string")
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	page, size := c.getPaginationData()

	status := c.Ctx.URLParamDefault("status", "all")

	taskCount, taskStatusList := c.Service.GetTaskPlayer(taskID, status, page, size)
	if taskStatusList == nil {
		taskStatusList = []services.TaskStatus{}
	}

	res := PlayerListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: taskCount,
		},
		Data: taskStatusList,
	}
	c.JSON(res)
	return iris.StatusOK
}

// GetByWechat 生成活动微信小程序码
func (c *TaskController) GetByWechat(id string) int {
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	url := c.Service.GetQRCode(taskID)
	c.JSON(struct {
		URL string `json:"url"`
	}{
		URL: url,
	})
	return iris.StatusOK
}

package controllers

import (
	"strconv"

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

// Post 添加任务
func (c *TaskController) Post() int {
	id := c.checkLogin()
	req := AddTaskReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)

	taskType := models.TaskType(req.Type)
	libs.Assert(taskType == models.TaskTypeInfo ||
		taskType == models.TaskTypeQuestionnaire ||
		taskType == models.TaskTypeRunning, "invalid_type", 400)

	libs.CheckReward(req.Reward, req.RewardObject, req.RewardValue)
	taskReward := models.RewardType(req.Reward)

	libs.Assert(req.Title != "", "invalid_title", 400)
	libs.Assert(req.Content != "", "invalid_content", 400)

	libs.CheckDateDuring(req.StartDate, req.EndDate)

	libs.Assert(req.MaxPlayer > 0, "invalid_max_player", 400)

	libs.Assert(len(req.Title) < 64, "title_too_long", 403)
	libs.Assert(len(req.Content) < 512, "content_too_long", 403)
	libs.Assert(len(req.RewardObject) < 32, "reward_object_too_long", 403)

	for _, l := range req.Location {
		libs.Assert(len(l) < 64, "location_too_long", 403)
	}

	for _, t := range req.Tags {
		libs.Assert(len(t) < 16, "tag_too_long", 403)
	}

	var images []primitive.ObjectID
	for _, file := range req.Images {
		fileID, err := primitive.ObjectIDFromHex(file)
		libs.AssertErr(err, "invalid_file", 400)
		images = append(images, fileID)
	}
	var attachments []primitive.ObjectID
	for _, attachment := range req.Attachment {
		fileID, err := primitive.ObjectIDFromHex(attachment)
		libs.AssertErr(err, "invalid_file", 400)
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
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "string")
	task := c.Service.GetTaskByID(_id, c.Session.GetString("id"))
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

	libs.Assert(len(req.Title) < 32, "title_too_long", 403)
	libs.Assert(len(req.Content) < 512, "content_too_long", 403)
	libs.Assert(len(req.RewardObject) < 32, "reward_object_too_long", 403)

	for _, l := range req.Location {
		libs.Assert(len(l) < 64, "location_too_long", 403)
	}

	for _, t := range req.Tags {
		libs.Assert(len(t) < 16, "tag_too_long", 403)
	}

	var images []primitive.ObjectID
	for _, file := range req.Images {
		fileID, err := primitive.ObjectIDFromHex(file)
		libs.AssertErr(err, "invalid_file", 400)
		images = append(images, fileID)
	}
	var attachments []primitive.ObjectID
	for _, attachment := range req.Attachment {
		fileID, err := primitive.ObjectIDFromHex(attachment)
		libs.AssertErr(err, "invalid_file", 400)
		attachments = append(attachments, fileID)
	}

	taskInfo := models.TaskSchema{
		Title:        req.Title,
		Content:      req.Content,
		Location:     req.Location,
		Tags:         req.Tags,
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
	pageStr := c.Ctx.URLParamDefault("page", "1")
	page, err := strconv.ParseInt(pageStr, 10, 64)
	libs.AssertErr(err, "invalid_page", 400)
	sizeStr := c.Ctx.URLParamDefault("size", "10")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	libs.AssertErr(err, "invalid_size", 400)

	sort := c.Ctx.URLParamDefault("sort", "new")
	taskType := c.Ctx.URLParamDefault("type", "all")
	status := c.Ctx.URLParamDefault("status", "wait")
	reward := c.Ctx.URLParamDefault("reward", "all")
	keyword := c.Ctx.URLParamDefault("keyword", "")
	user := c.Ctx.URLParamDefault("user", "")

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
		taskType, status, reward, keyword, user, c.Session.GetString("id"))

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

// PostByView 添加任务阅读量
func (c *TaskController) PostByView(id string) int {
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.AddView(taskID)
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

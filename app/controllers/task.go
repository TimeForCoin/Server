package controllers

import (
	"fmt"
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
)

type TaskController struct {
	BaseController
	Server services.TaskService
}

func BindTaskController(app *iris.Application) {
	taskService := services.GetServiceManger().Task

	taskRoute := mvc.New(app.Party("/tasks"))
	taskRoute.Register(taskService, getSession().Start)
	taskRoute.Handle(new(TaskController))
}

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

type AttachmentData struct {
	ID          primitive.ObjectID `json:"id"`
	Type        models.FileType    `json:"type"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Size        int64              `json:"size"`
	Time        int64              `json:"time"`
	Public      bool               `json:"public"`
}

type GetTaskInfoRes struct {
	*models.TaskSchema
	// 额外项
	Publisher  models.UserBaseInfo
	Attachment []AttachmentData
	Images     []string
	Like       bool
	// 排除项
	LikeID omit `json:"like_id,omitempty"` // 点赞用户ID
}

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
	libs.Assert(req.Title != "", "invalid_title", 400)

	libs.CheckDateDuring(req.StartDate, req.EndDate)

	libs.Assert(req.MaxPlayer > 0, "invalid_max_player", 400)

	var files []primitive.ObjectID
	req.Images = append(req.Images, req.Attachment...)
	for _, file := range req.Images {
		fileID, err := primitive.ObjectIDFromHex(file)
		libs.AssertErr(err, "invalid_file", 400)
		files = append(files, fileID)
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
	c.Server.AddTask(id, taskInfo, req.Publish)
	return iris.StatusOK
}
func (c *TaskController) GetBy(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "string")
	task, publisher, _, _ := c.Server.GetTaskByID(_id)

	isLike := false
	userID := c.Session.GetString("id")
	if userID != "" {
		_id, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			for _, likeUser := range task.LikeID {
				if likeUser == _id {
					isLike = true
				}
			}
		}
	}

	c.JSON(GetTaskInfoRes{
		TaskSchema: &task,
		Publisher:  publisher,
		Attachment: []AttachmentData{},
		Images:     []string{},
		Like:       isLike,
	})
	return iris.StatusOK
}
func (c *TaskController) PatchBy(id string) int {
	_ = c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	req := AddTaskReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.CheckReward(req.Reward, req.RewardObject, req.RewardValue)
	taskReward := models.RewardType(req.Reward)
	// TODO 检验用户权限

	// TODO 检查时间

	libs.AssertErr(err, "invalid_value", 400)
	taskInfo := models.TaskSchema{
		Title:        req.Title,
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
	c.Server.SetTaskInfo(taskID, taskInfo)
	return iris.StatusOK
}

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

type PaginationRes struct {
	Page  int
	Size  int
	Total int
}

type TasksListRes struct {
	Pagination PaginationRes
	Tasks      []GetTaskInfoRes
}

func (c *TaskController) Get() int {

	pageStr := c.Ctx.URLParamDefault("page", "1")
	page, err := strconv.Atoi(pageStr)
	libs.AssertErr(err, "invalid_page", 400)
	sizeStr := c.Ctx.URLParamDefault("size", "10")
	size, err := strconv.Atoi(sizeStr)
	libs.AssertErr(err, "invalid_size", 400)

	sort := c.Ctx.URLParamDefault("sort", "new")
	taskType := c.Ctx.URLParamDefault("type", "all")
	status := c.Ctx.URLParamDefault("status", "wait,run")
	reward := c.Ctx.URLParamDefault("reward", "all")
	userFilter := c.Ctx.URLParamDefault("user", "all")
	keyword := c.Ctx.URLParamDefault("keyword", "")

	fmt.Println(page)

	taskCount, tasksData := c.Server.GetTasks(page, size, sort,
		taskType, status, reward, userFilter, keyword)

	var tasks []GetTaskInfoRes
	//{
	//	TaskSchema: &task,
	//	Publisher:  publisher,
	//	Attachment: []AttachmentData{},
	//	Images:     []string{},
	//	Like:       isLike,
	//}

	userID := c.Session.GetString("id")
	for _, t := range tasksData {
		userService := services.GetServiceManger().User
		isLike := false
		if userID != "" {
			_id, err := primitive.ObjectIDFromHex(userID)
			if err == nil {
				for _, likeUser := range t.LikeID {
					if likeUser == _id {
						isLike = true
					}
				}
			}
		}


		tasks = append(tasks, GetTaskInfoRes{
			TaskSchema: &t,
			Publisher: userService.GetUserBaseInfo(t.Publisher),
			Attachment: []AttachmentData{},
			Images: []string{},
			Like: isLike,
		})
	}

	res := TasksListRes{
		Pagination: PaginationRes{
			Page:  page,
			Size:  size,
			Total: taskCount,
		},
		Tasks: tasks,
	}
	c.JSON(res)
	return iris.StatusOK
}

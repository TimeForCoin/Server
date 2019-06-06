package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Title        string               `json:"title"`
	Content      string               `json:"content"`
	Attachment   []primitive.ObjectID `json:"attachment"`
	Type         models.TaskType      `json:"type"`
	Reward       models.RewardType    `json:"reward"`
	RewardValue  float32              `json:"reward_value"`
	RewardObject string               `json:"reward_object"`
	Location     []string             `json:"location"`
	Tags         []string             `json:"tags"`
	TopTime      int64                `json:"top_time"`
	StartDate    int64                `json:"start_date"`
	EndDate      int64                `json:"end_date"`
	MaxPlayer    int64                `json:"max_player"`
	MaxFinish    int64                `json:"max_finish"`
	AutoAccept   bool                 `json:"auto_accept"`
	Publish      bool                 `json:"publish"`
}

type PublisherData struct {
	ID       primitive.ObjectID `json:"id"`
	Nickname string             `json:"nickname"`
	Avatar   string             `json:"avatar"`
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

type TaskData struct {
	*models.TaskSchema
	// 额外项
	Publisher  PublisherData
	Attachment []AttachmentData
	Like       bool
	// 排除项
	LikeID omit `json:"like_id,omitempty"` // 点赞用户ID
}

type GetTaskInfoByIDRes struct {
	Data *TaskData `json:"data"`
}

func (c *TaskController) PostTasks() int {
	id := primitive.NewObjectID().Hex()
	req := AddTaskReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	taskInfo := models.TaskSchema{
		Title:        req.Title,
		Type:         req.Type,
		Content:      req.Content,
		Attachment:   req.Attachment,
		Location:     req.Location,
		Tags:         req.Tags,
		TopTime:      req.TopTime,
		Reward:       req.Reward,
		RewardValue:  req.RewardValue,
		RewardObject: req.RewardObject,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		MaxPlayer:    req.MaxPlayer,
		MaxFinish:    req.MaxFinish,
		AutoAccept:   req.AutoAccept,
	}
	_, success := c.Server.AddTask(id, taskInfo)
	libs.Assert(success == true, "invalid_title", 403)
	return iris.StatusOK
}
func (c *TaskController) GetInfoBy(id string) int {
	// _ :=  c.checkLogin()
	_, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "string")
	task, publisher, attachments, err := c.Server.GetTaskByID(id)
	var attachmentsData []AttachmentData
	for _, attachment := range attachments {
		attachmentData := AttachmentData{
			ID:          attachment.ID,
			Type:        attachment.Type,
			Name:        attachment.Name,
			Description: attachment.Description,
			Size:        attachment.Size,
			Time:        attachment.Time,
			Public:      attachment.Public,
		}
		attachmentsData = append(attachmentsData, attachmentData)
	}
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
	c.JSON(GetTaskInfoByIDRes{
		Data: &TaskData{
			TaskSchema: &task,
			Publisher: PublisherData{
				ID:       publisher.ID,
				Nickname: publisher.Info.Nickname,
				Avatar:   publisher.Info.Avatar,
			},
			Attachment:   attachmentsData,
			Like:         isLike,
		},
	})
	return iris.StatusOK
}
func (c *TaskController) PatchInfoBy(id string) int {
	_ = primitive.NewObjectID().Hex()
	req := AddTaskReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	taskInfo := models.TaskSchema{
		Title:        req.Title,
		Type:         req.Type,
		Content:      req.Content,
		Attachment:   req.Attachment,
		Location:     req.Location,
		Tags:         req.Tags,
		TopTime:      req.TopTime,
		Reward:       req.Reward,
		RewardValue:  req.RewardValue,
		RewardObject: req.RewardObject,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		MaxPlayer:    req.MaxPlayer,
		MaxFinish:    req.MaxFinish,
		AutoAccept:   req.AutoAccept,
	}
	err = c.Server.SetTaskInfo(id, taskInfo)
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

type TasksListRes struct {
	TotalPages int `json:"total_pages"`
	Tasks      []models.TaskCard
}

func (c *TaskController) GetTasks() int {
	id := c.checkLogin()
	req := GetTasksReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.Assert(err == nil, "invalid_value", 400)
	totalPages, tasks := c.Server.GetTasks(id, req.Page, req.Size, req.Sort,
		req.Type, req.Status, req.Reward, req.User, req.Keyword)
	res := TasksListRes{
		TotalPages: totalPages,
		Tasks:      tasks,
	}
	c.JSON(res)
	return iris.StatusOK
}

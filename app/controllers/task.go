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
	taskService := services.NewTaskService()

	taskRoute := mvc.New(app.Party("/tasks"))
	taskRoute.Register(taskService, getSession().Start)
	taskRoute.Handle(new(TaskController))
}

type AddTaskReq struct {
	Title string `json:"title"`
	Content string `json:"content"`
	Attachment []primitive.ObjectID `json:"attachment"`
	Type models.TaskType `json:"type"`
	Reward models.RewardType `json:"reward"`
	RewardValue float32 `json:"reward_value"`
	RewardObject string `json:"reward_object"`
	Location []string `json:"location"`
	Tags []string `json:"tags"`
	TopTime int64 `json:"top_time"`
	StartDate int64 `json:"start_date"`
	EndDate int64 `json:"end_date"`
	MaxPlayer int64 `json:"max_player"`
	MaxFinish int64 `json:"max_finish"`
	AutoAccept bool `json:"auto_accept"`
	Publish bool `json:"publish"`
}

type PublisherData struct {
	ID primitive.ObjectID `json:"id"`
	Nickname string `json:"nickname"`
	Avatar string `json:"avatar"`
}

type AttachmentData struct {
	ID primitive.ObjectID `json:"id"`
	Type models.FileType `json:"type"`
	Name string `json:"name"`
	Description string `json:"description"`
	Size int64 `json:"size"`
	Time int64 `json:"time"`
	Public bool `json:"public"`
}



type TaskData struct {
	ID primitive.ObjectID `json:"id"`
	Publisher *PublisherData `json:"publisher"`
	Title string `json:"title"`
	Content string `json:"content"`
	Location []string `json:"location"`
	Tags []string `json:"tags"`
	TopTime int64 `json:"top_time"`
	Status models.TaskStatus `json:"status"`
	Type models.TaskType `json:"type"`
	Attachment []AttachmentData `json:"attachment"`
	Reward models.RewardType `json:"reward"`
	RewardValue float32 `json:"reward_value"`
	RewardObject string `json:"reward_object"`
	PublishDate int64 `json:"publish_date"`
	StartDate int64 `json:"start_date"`
	EndDate int64 `json:"end_date"`
	PlayerCount int64 `json:"player_count"`
	MaxPlayer int64 `json:"max_player"`
	MaxFinish int64 `json:"max_finish"`
	AutoAccept bool `json:"auto_accept"`
	CommentCount int64 `json:"comment_count"`
	ViewCount int64 `json:"view_count"`
	CollectCount int64 `json:"collect_count"`
	LikeCount int64 `json:"like_count"`
	Like bool `json:"like"`
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
	//_ :=  c.checkLogin()
	_, err := primitive.ObjectIDFromHex(id)
	libs.Assert(err == nil, "string")
	task, publisher, attachments, err := c.Server.GetTaskByID(id)
	attachmentsData := []AttachmentData{}
	for _,attachment := range attachments {
		attachmentData := AttachmentData {
			ID            :attachment.ID,         
			Type          :attachment.Type,      
			Name          :attachment.Name,       
			Description   :attachment.Description,
			Size          :attachment.Size,       
			Time          :attachment.Time,      
			Public        :attachment.Public,     
		}
		attachmentsData = append(attachmentsData, attachmentData)
	}
	c.JSON(GetTaskInfoByIDRes{
		Data: &TaskData{
			ID           :task.ID,
			Publisher    :&PublisherData {
				ID        :publisher.ID,
				Nickname  :publisher.Info.Nickname,
				Avatar    :publisher.Info.Avatar,
			},
			Title        :task.Title,
			Content      :task.Content,
			Location     :task.Location,    
			Tags         :task.Tags, 
			TopTime      :task.TopTime,      
			Status       :task.Status,
			Type         :task.Type,
			Attachment   :attachmentsData,
			Reward       :task.Reward,
			RewardValue  :task.RewardValue,      
			RewardObject :task.RewardObject,          
			PublishDate  :task.PublishDate,         
			StartDate    :task.StartDate,   
			EndDate      :task.EndDate, 
			PlayerCount  :task.PlayerCount, 
			MaxPlayer    :task.MaxPlayer,   
			MaxFinish    :task.MaxFinish,      
			AutoAccept   :task.AutoAccept,
			CommentCount :task.CommentCount,          
			ViewCount    :task.ViewCount,              
			CollectCount :task.CollectCount,                  
			LikeCount    :task.LikeCount,         
			Like         :false,
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
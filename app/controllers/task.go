package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

type TaskController struct {
	BaseController
	Server services.TaskService
}

func BindTaskController(app *iris.Application) {
	taskService := services.GetServiceManger().Task

	taskRoute := mvc.New(app.Party("/task"))
	taskRoute.Register(taskService, getSession().Start)
	taskRoute.Handle(new(TaskController))
}

type GetTasksReq struct {
	Page	int
	Size	int
	Sort	string
	Type	string
	Status	string
	Reward	string
	User	string
	Keyword	string
}

type TasksListRes struct {
	TotalPages	int `json:"total_pages"`
	Tasks		[]models.TaskCard
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
		Tasks: tasks,
	}
	c.JSON(res)
	return iris.StatusOK
}



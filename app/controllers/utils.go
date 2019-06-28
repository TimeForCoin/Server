package controllers

import (
	"strconv"

	"github.com/TimeForCoin/Server/app/services"
	"github.com/TimeForCoin/Server/app/utils"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskController 任务相关API
type UtilsController struct {
	BaseController
	Service services.UtilsService
}

// BindTaskController 绑定任务控制器
func BindUtilsController(app *iris.Application) {
	utilsService := services.GetServiceManger().Utils

	utilsRoute := mvc.New(app.Party("/utils"))
	utilsRoute.Register(utilsService, getSession().Start)
	utilsRoute.Handle(new(UtilsController))
}

type LogsRes struct {
	Count int64
	Data  []services.LogDetail
}

func (c *UtilsController) GetLogs() int {
	page, size := c.getPaginationData()

	user := c.Ctx.URLParamDefault("user", "me")
	logType := c.Ctx.URLParamDefault("type", "all")
	startDateStr := c.Ctx.URLParamDefault("start_date", "0")
	endDateStr := c.Ctx.URLParamDefault("end_date", "0")

	startDate, err := strconv.ParseInt(startDateStr, 10, 64)
	endDate, err := strconv.ParseInt(endDateStr, 10, 64)

	postUserID := c.checkLogin()
	var userID primitive.ObjectID
	if user == "me" {
		userID = postUserID
	} else {
		userID, err = primitive.ObjectIDFromHex(user)
		utils.AssertErr(err, "invalid_id", 400)
	}

	logCount, logData := c.Service.GetLogs(page, size, logType, userID, postUserID, startDate, endDate)

	res := LogsRes{
		Count: logCount,
		Data:  logData,
	}
	c.JSON(res)
	return iris.StatusOK
}

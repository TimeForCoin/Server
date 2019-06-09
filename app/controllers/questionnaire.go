package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuestionnaireController struct {
	BaseController
	Server services.QuestionnaireService
}

func BindQuestionnaireController(app *iris.Application) {
	questionnaireService := services.GetServiceManger().Questionnaire

	questionnaireRoute := mvc.New(app.Party("/questionnaires"))
	questionnaireRoute.Register(questionnaireService, getSession().Start)
	questionnaireRoute.Handle(new(QuestionnaireController))
}

type AddQuestionnaireReq struct {
	Title		string	`json:"title"`
	Description	string	`json:"description"`
	Anonymous	bool	`json:"anonymous"`
}

type QuestionsRes struct {
	Data	[]models.ProblemSchema	`json:"data"`
}

func (c *QuestionnaireController) Post(id string) int {
	userID := c.checkLogin()
	req := AddQuestionnaireReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	libs.Assert(req.Title != "", "invalid_title", 400)
	// TODO 未登录/登陆过期
	// TODO 权限不足

	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	questionnaire := models.QuestionnaireSchema{
		TaskID: taskID,
		Title: req.Title,
		Description: req.Description,
		Owner: userID.String(),
		Anonymous: req.Anonymous,
	}
	c.Server.AddQuestionnaire(questionnaire)
	return iris.StatusOK
}

func (c *QuestionnaireController) GetByID(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	// TODO 未登录/登陆过期

	questionnaireInfo := c.Server.GetQuestionnaireInfoByID(_id)
	c.JSON(questionnaireInfo)
	return iris.StatusOK
}

func (c *QuestionnaireController) PatchBy(id string) int {
	_ = c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	// TODO 未登录/登陆过期
	// TODO 权限不足

	req := AddQuestionnaireReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	questionnaireInfo := models.QuestionnaireSchema{
		TaskID:			taskID,
		Title:			req.Title,
		Description:	req.Description,
		Anonymous:		req.Anonymous,
	}
	c.Server.SetQuestionnaireInfo(questionnaireInfo)
	return iris.StatusOK
}

func (c *QuestionnaireController) GetQuestionsBy(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	// TODO 未登录/登陆过期
	// TODO 权限不足

	questions := c.Server.GetQuestionnaireQuestionsByID(_id)
	c.JSON(QuestionsRes{
		Data:	questions,
	})
	return iris.StatusOK
}

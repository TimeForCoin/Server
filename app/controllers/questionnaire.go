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

type Questions struct {
	Data	[]models.ProblemSchema	`json:"data"`
}

func (c *QuestionnaireController) Post(id string) int {
	userID := c.checkLogin()
	req := AddQuestionnaireReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	libs.Assert(req.Title != "", "invalid_title", 400)

	// TODO 未登录/登陆过期

	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	questionnaire := models.QuestionnaireSchema{
		TaskID: taskID,
		Title: req.Title,
		Description: req.Description,
		Owner: userID.String(),
		Anonymous: req.Anonymous,
	}
	c.Server.AddQuestionnaire(userID, questionnaire)
	return iris.StatusOK
}

func (c *QuestionnaireController) GetBy(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	questionnaireInfo := c.Server.GetQuestionnaireInfoByID(_id)
	c.JSON(questionnaireInfo)
	return iris.StatusOK
}

func (c *QuestionnaireController) PutBy(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	req := AddQuestionnaireReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	questionnaireInfo := models.QuestionnaireSchema{
		TaskID:			taskID,
		Title:			req.Title,
		Description:	req.Description,
		Anonymous:		req.Anonymous,
	}
	c.Server.SetQuestionnaireInfo(userID, questionnaireInfo)
	return iris.StatusOK
}

func (c *QuestionnaireController) GetByQuestion(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	questions := c.Server.GetQuestionnaireQuestionsByID(_id)
	c.JSON(Questions{
		Data:	questions,
	})
	return iris.StatusOK
}

func (c *QuestionnaireController) PostByQuestion(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	req := Questions{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	// TODO 未登录/登陆过期

	c.Server.SetQuestionnaireQuestions(userID, _id, req.Data)
	return iris.StatusOK
}

func (c *QuestionnaireController) GetByAnswer(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	c.Server.GetQuestionnaireAnswersByID(userID, _id)
	return iris.StatusOK
}

func (c * QuestionnaireController) PostByAnswer(id string) int {
	_ := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/权限过期

	req := models.StatisticsSchema{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	c.Server.AddAnswer(_id, req)
	return iris.StatusOK
}

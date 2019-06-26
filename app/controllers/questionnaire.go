package controllers

import (
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/TimeForCoin/Server/app/utils"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuestionnaireController 问卷相关API
type QuestionnaireController struct {
	BaseController
	Server services.QuestionnaireService
}

// BindQuestionnaireController 绑定问卷控制器
func BindQuestionnaireController(app *iris.Application) {
	questionnaireService := services.GetServiceManger().Questionnaire

	questionnaireRoute := mvc.New(app.Party("/questionnaires"))
	questionnaireRoute.Register(questionnaireService, getSession().Start)
	questionnaireRoute.Handle(new(QuestionnaireController))
}

// AddQuestionnaireReq 添加问卷请求
type AddQuestionnaireReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Anonymous   bool   `json:"anonymous"`
}

// QuestionsRes 问题请求
type QuestionsRes struct {
	Data []models.ProblemSchema `json:"problems"`
}

// PostBy 新建问卷
func (c *QuestionnaireController) PostBy(id string) int {
	userID := c.checkLogin()
	req := AddQuestionnaireReq{}
	err := c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)
	utils.Assert(req.Title != "", "invalid_title", 400)

	taskID, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)

	questionnaire := models.QuestionnaireSchema{
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		Owner:       userID,
		Anonymous:   req.Anonymous,
	}
	c.Server.AddQuestionnaire(questionnaire)
	return iris.StatusOK
}

// GetBy 获取问卷信息
func (c *QuestionnaireController) GetBy(id string) int {
	_id, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)

	questionnaireInfo := c.Server.GetQuestionnaireInfoByID(_id)
	c.JSON(questionnaireInfo)
	return iris.StatusOK
}

// PutBy 修改问卷信息
func (c *QuestionnaireController) PutBy(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)

	req := AddQuestionnaireReq{}
	err = c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)
	utils.Assert(req.Title != "", "invalid_title", 400)
	utils.Assert(req.Description != "", "invalid_description", 400)
	questionnaireInfo := models.QuestionnaireSchema{
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		Anonymous:   req.Anonymous,
	}
	c.Server.SetQuestionnaireInfo(userID, questionnaireInfo)
	return iris.StatusOK
}

// GetByQuestions 获取问卷问题
func (c *QuestionnaireController) GetByQuestions(id string) int {
	_id, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)
	questions := c.Server.GetQuestionnaireQuestionsByID(_id)
	if questions == nil {
		questions = []models.ProblemSchema{}
	}
	c.JSON(QuestionsRes{
		Data: questions,
	})
	return iris.StatusOK
}

// PostByQuestions 修改问卷问题
func (c *QuestionnaireController) PostByQuestions(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)
	req := QuestionsRes{}
	err = c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)

	c.Server.SetQuestionnaireQuestions(userID, _id, req.Data)
	return iris.StatusOK
}

// GetByAnswers 获取问卷答案数据
func (c *QuestionnaireController) GetByAnswers(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)
	res := c.Server.GetQuestionnaireAnswersByID(userID, _id)
	c.JSON(res)

	return iris.StatusOK
}

// PostAnswersReq 添加回答请求
type PostAnswersReq struct {
	Data []models.ProblemDataSchema `json:"data"`
}

// PostByAnswers 添加新回答
func (c *QuestionnaireController) PostByAnswers(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	utils.AssertErr(err, "invalid_id", 400)

	req := PostAnswersReq{}
	err = c.Ctx.ReadJSON(&req)
	utils.AssertErr(err, "invalid_value", 400)

	c.Server.AddAnswer(_id, userID, req.Data)
	return iris.StatusOK
}

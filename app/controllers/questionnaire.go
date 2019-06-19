package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
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

type QuestionsRes struct {
	Data []models.ProblemSchema `json:"data"`
}

// Post 新建问卷
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
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		Owner:       userID.String(),
		Anonymous:   req.Anonymous,
	}
	c.Server.AddQuestionnaire(questionnaire)
	return iris.StatusOK
}


// GetByID 获取问卷信息
func (c *QuestionnaireController) GetByID(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	questionnaireInfo := c.Server.GetQuestionnaireInfoByID(_id)
	c.JSON(questionnaireInfo)
	return iris.StatusOK
}


// PatchBy 修改问卷信息
func (c *QuestionnaireController) PutBy(id string) int {
	userID := c.checkLogin()
	taskID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	req := AddQuestionnaireReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	questionnaireInfo := models.QuestionnaireSchema{
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		Anonymous:   req.Anonymous,
	}
	c.Server.SetQuestionnaireInfo(userID, questionnaireInfo)
	return iris.StatusOK
}

// GetQuestionsBy 获取问卷问题
func (c *QuestionnaireController) GetQuestionsBy(id string) int {
	// _ :=  c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	questions := c.Server.GetQuestionnaireQuestionsByID(_id)
	c.JSON(QuestionsRes{
		Data:	questions,
	})
	return iris.StatusOK
}

// PostQuestionsBy 修改问卷问题
func (c *QuestionnaireController) PostQuestionsBy(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	req := QuestionsRes{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	// TODO 未登录/登陆过期

	c.Server.SetQuestionnaireQuestions(userID, _id, req.Data)
	return iris.StatusOK
}

// GetAnswersBy 获取问卷答案数据
func (c *QuestionnaireController) GetAnswersBy(id string) int {
	userID := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/登陆过期

	c.Server.GetQuestionnaireAnswersByID(userID, _id)
	return iris.StatusOK
}

// PostAnswerBy 添加新回答
func (c * QuestionnaireController) PostAnswerBy(id string) int {
	//_ := c.checkLogin()
	_id, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	// TODO 未登录/权限过期

	req := models.StatisticsSchema{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	c.Server.AddAnswer(_id, req)
	return iris.StatusOK
}

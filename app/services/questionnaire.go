package services

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuestionnaireService interface {
	AddQuestionnaire(userID primitive.ObjectID, info models.QuestionnaireSchema)
	SetQuestionnaireInfo(userID primitive.ObjectID, info models.QuestionnaireSchema)
	GetQuestionnaireInfoByID(id primitive.ObjectID) (detail QuestionnaireDetail)
	GetQuestionnaireQuestionsByID(id primitive.ObjectID) (questions []models.ProblemSchema)
	SetQuestionnaireQuestions(userID primitive.ObjectID, id primitive.ObjectID, questions []models.ProblemSchema)
	GetQuestionnaireAnswersByID(userID primitive.ObjectID, id primitive.ObjectID) (QuestionnaireStatisticsRes)
	AddAnswer(id primitive.ObjectID, statistics models.StatisticsSchema)
}

func newQuestionnaireService() QuestionnaireService {
	return &questionnaireService{
		model:		models.GetModel().Questionnaire,
		userModel:	models.GetModel().User,
		taskModel:	models.GetModel().Task,
		cacheModel: models.GetRedis().Cache,
	}
}

type questionnaireService struct {
	model 		*models.QuestionnaireModel
	userModel	*models.UserModel
	taskModel	*models.TaskModel
	cacheModel	*models.CacheModel
}

type OwnerInfo	struct{
	ID			string	`json:"id"`
	NickName	string	`json:"nickname"`
	Avatar		string	`json:"avatar"`
}

type QuestionnaireDetail struct {
	TaskID			string		`json:"id"`
	Title			string		`json:"title"`
	Owner			OwnerInfo	`json:"owner"`
	Description		string		`json:"description"`
	Anonymous		bool		`json:"anonymous"`
	QuestionCount	int			`json:"question_count"`
	Answer			int			`json:"answer"`
}

type QuestionnaireStatisticsRes struct {
	Count	int							`json:"count"`
	Data	[]models.StatisticsSchema	`json:"data"`
}

func (s *questionnaireService) AddQuestionnaire(userID primitive.ObjectID, info models.QuestionnaireSchema) {
	task, err := s.taskModel.GetTaskByID(info.TaskID)
	libs.AssertErr(err, "faked_task", 400)
	libs.Assert(task.Type == models.TaskTypeQuestionnaire, "not_allow", 403)
	libs.Assert(task.Publisher == userID, "permission_deny", 403)

	err = s.model.AddQuestionnaire(info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	err = s.model.SetQuestionnaireInfoByID(info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

func (s *questionnaireService) SetQuestionnaireInfo(userID primitive.ObjectID, info models.QuestionnaireSchema) {
	task, err := s.taskModel.GetTaskByID(info.TaskID)
	libs.AssertErr(err, "faked_task", 400)
	libs.Assert(task.Status == models.TaskStatusDraft, "not_allow", 403)

	libs.Assert(task.Publisher == userID, "permission_deny", 403)

	err = s.model.SetQuestionnaireInfoByID(info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

func (s *questionnaireService) GetQuestionnaireInfoByID(id primitive.ObjectID) (detail QuestionnaireDetail) {
	questionnaire, err := s.model.GetQuestionnaireInfoByID(id)
	libs.AssertErr(err, "faked_task", 400)
	ownerID, err := primitive.ObjectIDFromHex(questionnaire.Owner)
	libs.AssertErr(err, "invalid_id", 400)
	owner, err := s.cacheModel.GetUserBaseInfo(ownerID)
	libs.AssertErr(err, "faked_task", 400)
	detail = QuestionnaireDetail{
		TaskID:		questionnaire.TaskID.String(),
		Title:		questionnaire.Title,
		Owner: 		OwnerInfo{
			ID: questionnaire.Owner,
			NickName: owner.Nickname,
			Avatar: owner.Avatar,
		},
		Description:	questionnaire.Description,
		Anonymous:		questionnaire.Anonymous,
		QuestionCount:	len(questionnaire.Problems),
		Answer:			len(questionnaire.Data),
	}
	return
}

func (s *questionnaireService) GetQuestionnaireQuestionsByID(id primitive.ObjectID) (questions []models.ProblemSchema) {
	questions, err := s.model.GetQuestionnaireQuestionsByID(id)
	libs.AssertErr(err, "faked_task", 400)
	return
}

func (s *questionnaireService) SetQuestionnaireQuestions(userID primitive.ObjectID, id primitive.ObjectID, questions []models.ProblemSchema) {
	task, err := s.taskModel.GetTaskByID(id)
	libs.AssertErr(err, "faked_task", 400)
	libs.Assert(task.Publisher == userID, "permission_deny", 403)
	libs.Assert(task.Status == models.TaskStatusDraft, "not_allow", 403)

	err = s.model.SetQuestionnaireQuestionsByID(id, questions)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

func (s *questionnaireService) GetQuestionnaireAnswersByID(userID primitive.ObjectID, id primitive.ObjectID) (QuestionnaireStatisticsRes) {
	task, err := s.taskModel.GetTaskByID(id)
	libs.AssertErr(err, "faked_task", 400)
	libs.Assert(task.Publisher == userID, "permission_deny", 403)

	statistics, err := s.model.GetQuestionnaireAnswersByID(id)
	libs.AssertErr(err, "faked_task", 400)

	return QuestionnaireStatisticsRes{
		Count: len(statistics),
		Data:  statistics,
	}
}

func (s *questionnaireService) AddAnswer(id primitive.ObjectID, statistics models.StatisticsSchema) {
	task, err := s.taskModel.GetTaskByID(id)
	libs.AssertErr(err, "faked_task", 400)
	libs.Assert(task.Status == models.TaskStatusWait, "not_allow", 403)

	err = s.model.AddAnswer(id, statistics)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}
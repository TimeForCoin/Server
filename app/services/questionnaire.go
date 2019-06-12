package services

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuestionnaireService 问卷相关服务
type QuestionnaireService interface {
	AddQuestionnaire(info models.QuestionnaireSchema)
	SetQuestionnaireInfo(info models.QuestionnaireSchema)
	GetQuestionnaireInfoByID(id primitive.ObjectID) (detail QuestionnaireDetail)
	GetQuestionnaireQuestionsByID(id primitive.ObjectID) (questions []models.ProblemSchema)
}

func newQuestionnaireService() QuestionnaireService {
	return &questionnaireService{
		model:      models.GetModel().Questionnaire,
		userModel:  models.GetModel().User,
		cacheModel: models.GetRedis().Cache,
	}
}

type questionnaireService struct {
	model      *models.QuestionnaireModel
	userModel  *models.UserModel
	cacheModel *models.CacheModel
}

// QuestionnaireDetail 问卷详情信息
type QuestionnaireDetail struct {
	TaskID        string              `json:"id"`
	Title         string              `json:"title"`
	Owner         models.UserBaseInfo `json:"owner"`
	Description   string              `json:"description"`
	Anonymous     bool                `json:"anonymous"`
	QuestionCount int                 `json:"question_count"`
	Answer        int                 `json:"answer"`
}

// AddQuestionnaire 添加问卷
func (s *questionnaireService) AddQuestionnaire(info models.QuestionnaireSchema) {
	_, err := s.model.AddQuestionnaire(info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	err = s.model.SetQuestionnaireInfoByID(info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

// SetQuestionnaireInfo 设置问卷信息
func (s *questionnaireService) SetQuestionnaireInfo(info models.QuestionnaireSchema) {
	err := s.model.SetQuestionnaireInfoByID(info)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
}

// GetQuestionnaireInfoByID 获取问卷信息
func (s *questionnaireService) GetQuestionnaireInfoByID(id primitive.ObjectID) (detail QuestionnaireDetail) {
	questionnaire, err := s.model.GetQuestionnaireInfoByID(id)
	libs.AssertErr(err, "faked_task", 400)
	ownerID, err := primitive.ObjectIDFromHex(questionnaire.Owner)
	libs.AssertErr(err, "invalid_id", 400)
	owner, err := s.cacheModel.GetUserBaseInfo(ownerID)
	libs.AssertErr(err, "faked_task", 400)
	detail = QuestionnaireDetail{
		TaskID:        questionnaire.TaskID.String(),
		Title:         questionnaire.Title,
		Owner:         owner,
		Description:   questionnaire.Description,
		Anonymous:     questionnaire.Anonymous,
		QuestionCount: len(questionnaire.Problems),
		Answer:        len(questionnaire.Data),
	}
	return
}

// GetQuestionnaireQuestionsByID 获取问卷问题
func (s *questionnaireService) GetQuestionnaireQuestionsByID(id primitive.ObjectID) (questions []models.ProblemSchema) {
	questions, err := s.model.GetQuestionnaireQuestionsByID(id)
	libs.AssertErr(err, "faked_task", 400)
	return
}

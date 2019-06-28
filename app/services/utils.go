package services

import (
	"strings"

	"github.com/TimeForCoin/Server/app/utils"

	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskService 任务服务
type UtilsService interface {
	GetLogs(page, size int64, logsType string, userID, postUserID primitive.ObjectID, startDate, endDate int64) (logsCount int64, logs []LogDetail)
}

func newUtilsService() UtilsService {
	return &utilsService{
		model:           models.GetModel().Log,
		userModel:       models.GetModel().User,
		fileModel:       models.GetModel().File,
		cache:           models.GetRedis().Cache,
		setModel:        models.GetModel().Set,
		taskStatusModel: models.GetModel().TaskStatus,
		messageModel:    models.GetModel().Message,
	}
}

type utilsService struct {
	model           *models.LogModel
	userModel       *models.UserModel
	fileModel       *models.FileModel
	cache           *models.CacheModel
	setModel        *models.SetModel
	taskStatusModel *models.TaskStatusModel
	messageModel    *models.MessageModel
}

// TaskDetail 任务数据
type LogDetail struct {
	*models.LogSchema
	// 排除项
	ID omit `json:"_id,omitempty"`
}

func (s *utilsService) GetLogs(page, size int64, logsType string, userID, postUserID primitive.ObjectID,
	startDate, endDate int64) (logsCount int64, logs []LogDetail) {

	var logTypes []models.LogType

	user, err := s.userModel.GetUserByID(postUserID)
	isAdmin := user.Data.Type == models.UserTypeAdmin || user.Data.Type == models.UserTypeRoot

	split := strings.Split(logsType, ",")
	for _, str := range split {
		if str == "all" && isAdmin {
			logTypes = []models.LogType{models.LogTypeMoney, models.LogTypeValue, models.LogTypeCredit, models.LogTypeLogin, models.LogTypeClear, models.LogTypeStart, models.LogTypeError}
			break
		} else if str == "all" {
			logTypes = []models.LogType{models.LogTypeMoney, models.LogTypeValue, models.LogTypeLogin}
			break
		}
		logTypes = append(logTypes, models.LogType(str))
	}

	for _, logType := range logTypes {
		if logType == models.LogTypeCredit || logType == models.LogTypeClear || logType == models.LogTypeStart || logType == models.LogTypeError {
			utils.Assert(isAdmin, "permission_deny", 403)
		}
	}

	log, logsCount, err := s.model.GetLog(userID, logTypes, startDate, endDate, (page-1)*size, size)
	utils.AssertErr(err, "", iris.StatusInternalServerError)

	for i := range log {
		var logDetail LogDetail
		logDetail.LogSchema = &log[i]
		logs = append(logs, logDetail)
	}
	return

}

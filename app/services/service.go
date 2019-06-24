package services

var service *ServiceManger

// ServiceManger 服务管理器
type ServiceManger struct {
	User          UserService
	Article		  ArticleService
	Task          TaskService
	File          FileService
	Questionnaire QuestionnaireService
	Comment       CommentService
	Message       MessageService
}

// GetServiceManger 获取服务管理器
func GetServiceManger() *ServiceManger {
	if service == nil {
		service = &ServiceManger{
			User:          newUserService(),
			Article:	   newArticleService(),
			Task:          newTaskService(),
			File:          newFileService(),
			Questionnaire: newQuestionnaireService(),
			Comment:       newCommentService(),
			Message:       newMessageService(),
		}
	}
	return service
}

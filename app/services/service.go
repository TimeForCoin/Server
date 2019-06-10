package services

var service *ServiceManger

type ServiceManger struct {
	User          UserService
	Task          TaskService
	File          FileService
	Questionnaire QuestionnaireService
	Comment       CommentService
}

func GetServiceManger() *ServiceManger {
	if service == nil {
		service = &ServiceManger{
			User:          newUserService(),
			Task:          newTaskService(),
			File:          newFileService(),
			Questionnaire: newQuestionnaireService(),
			Comment:       newCommentService(),
		}
	}
	return service
}

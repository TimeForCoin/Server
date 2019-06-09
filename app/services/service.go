package services

var service *ServiceManger

type ServiceManger struct {
	User UserService
	Task TaskService
	File FileService
	Comment CommentService
}

func GetServiceManger() *ServiceManger {
	if service == nil {
		service = &ServiceManger{
			User: newUserService(),
			Task: newTaskService(),
			File: newFileService(),
			Comment: newCommentService(),
		}
	}
	return service
}
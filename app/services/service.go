package services

var service *ServiceManger

type ServiceManger struct {
	User UserService
	Task TaskService
	File FileService
}

func GetServiceManger() *ServiceManger {
	if service == nil {
		service = &ServiceManger{
			User: newUserService(),
			Task: newTaskService(),
			File: newFileService(),
		}
	}
	return service
}
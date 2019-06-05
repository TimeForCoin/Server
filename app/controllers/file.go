package controllers

import (
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

type FileController struct {
	BaseController
	Server services.FileService
}

func BindFileController(app *iris.Application) {
	fileService := services.GetServiceManger().File

	fileRoute := mvc.New(app.Party("/file"))
	fileRoute.Register(fileService, getSession().Start)
	fileRoute.Handle(new(FileController))
}

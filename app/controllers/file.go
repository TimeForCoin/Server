package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileController 文件相关API
type FileController struct {
	BaseController
	Service services.FileService
}

// BindFileController 绑定文件控制器
func BindFileController(app *iris.Application) {
	fileService := services.GetServiceManger().File

	fileRoute := mvc.New(app.Party("/file"))
	fileRoute.Register(fileService, getSession().Start)
	fileRoute.Handle(new(FileController))
}

// PostFileRes 新建文件请求
type PostFileRes struct {
	ID string `json:"id"`
}

// Post 新建文件
func (c *FileController) Post() int {
	id := c.checkLogin()

	file, head, err := c.Ctx.FormFile("data")
	//noinspection GoUnhandledErrorResult
	defer file.Close()
	libs.AssertErr(err, "invalid_data", 400)

	owner := c.Ctx.FormValueDefault("owner", "")
	libs.Assert(owner == "task" || owner == "user", "invalid_owner", 400)
	fileType := c.Ctx.FormValueDefault("type", "")
	libs.Assert(fileType == "image" || fileType == "file", "invalid_type", 400)

	name := c.Ctx.FormValueDefault("name", head.Filename)
	description := c.Ctx.FormValueDefault("description", "")
	publicStr := c.Ctx.FormValueDefault("public", "false")
	public := false
	if publicStr == "true" {
		public = true
	}

	fileID := c.Service.AddFile(file, *head, models.FileType(fileType),
		models.OwnerType(owner), id,
		name, description, public)

	c.JSON(&PostFileRes{
		ID: fileID.Hex(),
	})

	return iris.StatusOK
}

// PutFileReq 更新附件信息请求
type PutFileReq struct {
	Name        string
	Description string
	Public      bool
}

// PutBy 更新附件信息
func (c *FileController) PutBy(id string) int {
	userID := c.checkLogin()

	fileID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	req := PutFileReq{}
	libs.AssertErr(c.Ctx.ReadJSON(&req), "invalid_value", 400)

	c.Service.UpdateFileInfo(fileID, userID, req.Name, req.Description, req.Public)

	return iris.StatusOK
}

// Delete 清除未使用文件
func (c *FileController) Delete() int {
	// userID := c.checkLogin()
	return iris.StatusOK
}

package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
)

type CommentController struct {
	BaseController
	Service services.CommentService
}

func BindCommentController(app *iris.Application) {
	fileRoute := mvc.New(app.Party("/comments"))
	fileRoute.Register(services.GetServiceManger().Comment , getSession().Start)
	fileRoute.Handle(new(CommentController))
}

func (c *CommentController) GetBy(id string) int {
	pageStr := c.Ctx.URLParamDefault("page", "1")
	page, err := strconv.ParseInt(pageStr, 10 ,64)
	libs.AssertErr(err, "invalid_page", 400)
	sizeStr := c.Ctx.URLParamDefault("size", "10")
	size, err := strconv.ParseInt(sizeStr, 10 ,64)
	libs.AssertErr(err, "invalid_size", 400)

	sort := c.Ctx.URLParamDefault("sort", "new")

	contentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	res := c.Service.GetComments(contentID, c.Session.GetString("id"),  page, size, sort)
	c.JSON(res)
	return iris.StatusOK
}

type PostCommentReq struct {
	Type string
	Content string
}

func (c *CommentController) PostBy(id string) int {
	userID := c.checkLogin()

	req := PostCommentReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	contentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	libs.Assert(len(req.Content) < 128, "content_too_long", 403)

	if req.Type == "task" {
		c.Service.AddCommentForTask(userID, contentID, req.Content)
	} else if req.Type == "comment" {
		c.Service.AddCommentForComment(userID, contentID, req.Content)
	} else {
		libs.Assert(false, "invalid_type", 400)
	}

	return iris.StatusOK
}

func (c *CommentController) DeleteBy(id string) int {
	userID := c.checkLogin()

	commentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	c.Service.RemoveComment(userID, commentID)

	return iris.StatusOK
}

func (c *CommentController) PostByLike(id string) int {
	userID := c.checkLogin()

	commentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	c.Service.ChangeLike(userID, commentID, true)

	return iris.StatusOK
}

func (c *CommentController) DeleteByLike(id string) int {
	userID := c.checkLogin()

	commentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	c.Service.ChangeLike(userID, commentID, false)

	return iris.StatusOK
}
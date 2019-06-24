package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentController 评论相关API
type CommentController struct {
	BaseController
	Service services.CommentService
}

// BindCommentController 绑定评论API控制器
func BindCommentController(app *iris.Application) {
	fileRoute := mvc.New(app.Party("/comments"))
	fileRoute.Register(services.GetServiceManger().Comment, getSession().Start)
	fileRoute.Handle(new(CommentController))
}

// GetBy 获取相关内容的评论
func (c *CommentController) GetBy(id string) int {
	page, size := c.getPaginationData()

	sort := c.Ctx.URLParamDefault("sort", "new")

	contentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	res := c.Service.GetComments(contentID, c.Session.GetString("id"), page, size, sort)
	c.JSON(struct {
		Pagination PaginationRes
		Data       []services.CommentData
	}{
		Pagination: PaginationRes{
			Page: page,
			Size: size,
		},
		Data: res,
	})
	return iris.StatusOK
}

// PostCommentReq 添加评论请求
type PostCommentReq struct {
	Type    string
	Content string
}

// PostBy 添加评论
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

// DeleteBy 删除评论
func (c *CommentController) DeleteBy(id string) int {
	userID := c.checkLogin()

	commentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	c.Service.RemoveComment(userID, commentID)

	return iris.StatusOK
}

// PostByLike 评论点赞
func (c *CommentController) PostByLike(id string) int {
	userID := c.checkLogin()

	commentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	c.Service.ChangeLike(userID, commentID, true)

	return iris.StatusOK
}

// DeleteByLike 取消点赞
func (c *CommentController) DeleteByLike(id string) int {
	userID := c.checkLogin()

	commentID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	c.Service.ChangeLike(userID, commentID, false)

	return iris.StatusOK
}

package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArticleController struct {
	BaseController
	Service services.ArticleService
}

func BindArticleController(app *iris.Application) {
	articleService := services.GetServiceManger().Article

	articleRoute := mvc.New(app.Party("/article"))
	articleRoute.Register(articleService, getSession().Start)
	articleRoute.Handle(new(ArticleController))
}

// GetArticlesRes 公告文章列表数据
type ArticlesListRes struct {
	Pagination	PaginationRes
	Data		[]services.ArticleBrief
}

// ArticleReq 添加或修改公告文章请求
type ArticleReq struct {
	Title		string		`json:"title"`
	Content		string		`json:"content"`
	Publisher	string		`json:"publisher"`
	Images		[]string	`json:"images"`
}

// Get 获取公告文章列表
func (c *ArticleController) Get() int {
	page, size := c.getPaginationData()

	count, articles := c.Service.GetArticles(page, size)
	if articles == nil {
		articles = []services.ArticleBrief{}
	}
	res := ArticlesListRes{
		Pagination:	PaginationRes{
			Page:	page,
			Size:	size,
			Total:	count,
		},
		Data:	articles,
	}
	c.JSON(res)

	return iris.StatusOK
}

// Post 添加公告文章
func (c *ArticleController) Post() int {
	id := c.checkLogin()
	req := ArticleReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	libs.Assert(req.Title != "", "invalid_title", 400)
	libs.Assert(req.Content != "", "invalid_content", 400)
	libs.Assert(req.Publisher != "", "invalid_publisher", 400)

	var images []primitive.ObjectID
	for _, file := range req.Images {
		fileID, err := primitive.ObjectIDFromHex(file)
		libs.AssertErr(err, "invalid_file", 400)
		images = append(images, fileID)
	}

	c.Service.AddArticle(id, req.Title, req.Content, req.Publisher, images)
	return iris.StatusOK
}

// GetBy 根据ID获取公告文章详情
func (c *ArticleController) GetBy(id string) int {
	articleID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.GetArticleByID(articleID)
	return iris.StatusOK
}

// SetBy 根据ID修改公告文章详情
func (c *ArticleController) SetBy(id string) int {
	userID := c.checkLogin()
	articleID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	req := ArticleReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	var imageIDs []primitive.ObjectID
	for _, i := range req.Images {
		imageID, err := primitive.ObjectIDFromHex(i)
		libs.AssertErr(err, "invalid_value", 400)
		imageIDs = append(imageIDs, imageID)
	}
	c.Service.SetArticleByID(userID, articleID, req.Title, req.Content, req.Publisher, imageIDs)
	return iris.StatusOK
}

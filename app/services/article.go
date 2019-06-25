package services

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/models"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleService 公告服务
type ArticleService interface {
	GetArticles(page, size int64) (int64, []ArticleBrief)
	AddArticle(userID primitive.ObjectID, title string, content string, publisher string, images []primitive.ObjectID) primitive.ObjectID
	GetArticleByID(id primitive.ObjectID) ArticleDetail
	SetArticleByID(userID primitive.ObjectID, id primitive.ObjectID, title, content, publisher string, images []primitive.ObjectID)
}

func newArticleService() ArticleService {
	return &articleService{
		model:     models.GetModel().Article,
		userModel: models.GetModel().User,
		fileModel: models.GetModel().File,
	}
}

type articleService struct {
	model     *models.ArticleModel
	userModel *models.UserModel
	fileModel *models.FileModel
}

// ArticleBrief 公告简介
type ArticleBrief struct {
	ID        string
	Title     string
	Publisher string
	Images    []ImagesData
}

// ArticleDetail 公告详情
type ArticleDetail struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // ID
	ViewCount int64              `bson:"view_count"`    // 文章阅读数
	Title     string             // 文章标题
	Content   string             // 文章内容
	Publisher string             // 发布者名字
	Date      int64              // 发布时间
	Images    []ImagesData       // 首页图片
}

// GetArticles 获取公告列表
func (s *articleService) GetArticles(page, size int64) (count int64, articleList []ArticleBrief) {
	articles, count, err := s.model.GetArticles((page-1)*size, size)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	for _, article := range articles {
		var imagesData []ImagesData
		images, err := s.fileModel.GetFileByContent(article.ID, models.FileImage)
		libs.AssertErr(err, "", iris.StatusInternalServerError)
		for _, image := range images {
			imagesData = append(imagesData, ImagesData{
				ID:  image.ID.Hex(),
				URL: image.URL,
			})
		}
		brief := ArticleBrief{
			ID:        article.ID.Hex(),
			Title:     article.Title,
			Publisher: article.Publisher,
			Images:    imagesData,
		}
		articleList = append(articleList, brief)
	}
	return
}

// AddArticle 添加公告
func (s *articleService) AddArticle(userID primitive.ObjectID, title string, content string, publisher string, images []primitive.ObjectID) primitive.ObjectID {
	user, err := s.userModel.GetUserByID(userID)
	libs.AssertErr(err, "invalid_session", 400)
	libs.Assert(user.Data.Type == models.UserTypeAdmin || user.Data.Type == models.UserTypeRoot,
		"permission_deny", 500)

	articleID := primitive.NewObjectID()

	var files []FileBaseInfo
	for _, image := range images {
		files = append(files, FileBaseInfo{
			ID:   image,
			Type: models.FileImage,
		})
	}
	GetServiceManger().File.BindFilesToTask(userID, articleID, files)

	id, err := s.model.AddArticle(articleID, title, content, publisher, images)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	return id
}

// GetArticleByID 根据ID获取公告详情
func (s *articleService) GetArticleByID(id primitive.ObjectID) ArticleDetail {
	article, err := s.model.GetArticleByID(id)
	libs.AssertErr(err, "faked_article", 403)
	var imagesData []ImagesData
	images, err := s.fileModel.GetFileByContent(id, models.FileImage)
	libs.AssertErr(err, "", iris.StatusInternalServerError)
	for _, image := range images {
		imagesData = append(imagesData, ImagesData{
			ID:  image.ID.Hex(),
			URL: image.URL,
		})
	}
	return ArticleDetail{
		ID:        id,
		ViewCount: article.ViewCount,
		Title:     article.Title,
		Content:   article.Content,
		Publisher: article.Publisher,
		Date:      article.Date,
		Images:    imagesData,
	}
}

// SetArticleByID 根据ID修改公告文章
func (s *articleService) SetArticleByID(userID primitive.ObjectID, id primitive.ObjectID, title, content, publisher string, images []primitive.ObjectID) {
	user, err := s.userModel.GetUserByID(userID)
	libs.AssertErr(err, "invalid_session", 400)
	libs.Assert(user.Data.Type == models.UserTypeAdmin || user.Data.Type == models.UserTypeRoot,
		"permission_deny", 500)

	for _, imageID := range images {
		_, err := s.fileModel.GetFile(imageID)
		libs.AssertErr(err, "faked_image", 403)
	}

	var toRemove []primitive.ObjectID
	var imageFiles []FileBaseInfo
	if len(images) > 0 {
		oldImages, err := s.fileModel.GetFileByContent(id, models.FileImage)
		libs.AssertErr(err, "", 500)
		for _, image := range images {
			exist := false
			for _, old := range oldImages {
				if old.ID == image {
					exist = true
					break
				}
			}
			if !exist {
				imageFiles = append(imageFiles, FileBaseInfo{
					ID:   image,
					Type: models.FileImage,
				})
			}
		}
		for _, image := range oldImages {
			exist := false
			for _, file := range imageFiles {
				if file.ID == image.ID {
					exist = true
					break
				}
			}
			if !exist {
				toRemove = append(toRemove, image.ID)
			}
		}
	}

	GetServiceManger().File.BindFilesToTask(userID, id, imageFiles)

	err = s.model.SetArticleByID(id, title, content, publisher, images)
	libs.AssertErr(err, "", iris.StatusInternalServerError)

	// 删除无用文件
	for _, file := range toRemove {
		GetServiceManger().File.RemoveFile(file)
	}
}

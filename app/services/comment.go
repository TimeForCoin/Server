package services

import (
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentService 评论服务
type CommentService interface {
	AddCommentForTask(userID, taskID primitive.ObjectID, content string)
	AddCommentForComment(userID, commentID primitive.ObjectID, content string)
	RemoveComment(userID, commentID primitive.ObjectID)
	ChangeLike(userID, commentID primitive.ObjectID, like bool)
	GetComments(contentID primitive.ObjectID, userID string, page, size int64, sort string) []CommentData
}

// NewUserService 初始化
func newCommentService() CommentService {
	return &commentService{
		model:     models.GetModel().Comment,
		taskModel: models.GetModel().Task,
		setModel:  models.GetModel().Set,
		cache:     models.GetRedis().Cache,
	}
}

type commentService struct {
	model     *models.CommentModel
	taskModel *models.TaskModel
	setModel  *models.SetModel
	cache     *models.CacheModel
}

// AddCommentForTask 为任务添加评论
func (s *commentService) AddCommentForTask(userID, taskID primitive.ObjectID, content string) {
	task, err := s.taskModel.GetTaskByID(taskID)
	utils.AssertErr(err, "faked_content", 403)
	utils.Assert(task.Status != models.TaskStatusDraft, "not_allow_status", 403)
	err = s.model.AddComment(taskID, task.Publisher, userID, content, false)
	utils.AssertErr(err, "", 500)
	err = s.taskModel.InsertCount(taskID, models.CommentCount, 1)
	utils.AssertErr(err, "", 500)
}

// AddCommentForComment 为评论添加回复
func (s *commentService) AddCommentForComment(userID, commentID primitive.ObjectID, content string) {
	comment, err := s.model.GetCommentByID(commentID)
	utils.AssertErr(err, "faked_content", 403)
	utils.Assert(comment.IsReply == false, "faked_content", 403)
	err = s.model.AddComment(commentID, comment.UserID, userID, content, true)
	utils.AssertErr(err, "", 500)
	err = s.model.InsertCount(commentID, models.ReplyCount, 1)
	utils.AssertErr(err, "", 500)
}

// RemoveComment 删除评论
func (s *commentService) RemoveComment(userID, commentID primitive.ObjectID) {
	comment, err := s.model.GetCommentByID(commentID)
	utils.AssertErr(err, "faked_comment", 403)
	if comment.UserID != userID {
		userInfo, err := s.cache.GetUserBaseInfo(userID)
		utils.AssertErr(err, "", 500)
		utils.Assert(userInfo.Type == models.UserTypeAdmin || userInfo.Type == models.UserTypeRoot, "permission_deny", 403)
	}
	err = s.model.RemoveContentByID(commentID)
	utils.AssertErr(err, "", 500)
}

// ChangeLike 改变点赞状态
func (s *commentService) ChangeLike(userID, commentID primitive.ObjectID, like bool) {
	comment, err := s.model.GetCommentByID(commentID)
	utils.AssertErr(err, "faked_comment", 403)
	utils.Assert(comment.IsDelete == false, "deleted_comment", 403)
	if like {
		err := s.setModel.AddToSet(userID, commentID, models.SetOfLikeComment)
		utils.AssertErr(err, "exist_like", 403)
		err = s.model.InsertCount(commentID, models.LikeCount, 1)
		utils.AssertErr(err, "", 500)
	} else {
		err := s.setModel.RemoveFromSet(userID, commentID, models.SetOfLikeComment)
		utils.AssertErr(err, "faked_like", 403)
		err = s.model.InsertCount(commentID, models.LikeCount, -1)
		utils.AssertErr(err, "", 500)
	}
	err = s.cache.WillUpdate(userID, models.KindOfLikeComment)
	utils.AssertErr(err, "", 500)
}

// CommentWithUserInfo 带用户信息的评论数据
type CommentWithUserInfo struct {
	*models.CommentSchema
	// 额外项
	Own   models.UserBaseInfo
	User  models.UserBaseInfo
	Liked bool
	// 排除项
	ContentOwn omit `json:"content_own,omitempty"`
	UserID     omit `json:"user_id,omitempty"`
	IsReply    omit `json:"is_reply,omitempty"`
}

// CommentData 评论数据
type CommentData struct {
	*CommentWithUserInfo
	Reply []CommentWithUserInfo
}

// GetComments 获取评论列表
func (s *commentService) GetComments(contentID primitive.ObjectID, userID string, page, size int64, sort string) []CommentData {
	sortRule := bson.M{}
	if sort == "new" {
		sortRule["time"] = -1
	} else {
		sortRule["like_count"] = -1
	}
	comments, err := s.model.GetCommentsByContent(contentID, page, size, sortRule)
	utils.AssertErr(err, "faked_content", 403)

	if len(comments) == 0 {
		return []CommentData{}
	}
	var res []CommentData
	ownInfo, err := s.cache.GetUserBaseInfo(comments[0].ContentOwn)
	utils.AssertErr(err, "", 500)
	for i, c := range comments {
		comment := CommentData{}
		comment.CommentWithUserInfo = &CommentWithUserInfo{}
		comment.CommentSchema = &comments[i]
		comment.Own = ownInfo
		userInfo, err := s.cache.GetUserBaseInfo(c.UserID)
		utils.AssertErr(err, "", 500)
		comment.User = userInfo
		if userID != "" && c.IsDelete == false {
			_id, err := primitive.ObjectIDFromHex(userID)
			utils.AssertErr(err, "", 500)
			comment.Liked = s.cache.IsLikeComment(_id, c.ID)
		}
		if !c.IsReply && c.ReplyCount > 0 {
			// 默认显示最先5条回复
			replies, err := s.model.GetCommentsByContent(c.ID, 1, 5, bson.M{"time": 1})
			utils.AssertErr(err, "", 500)
			for j, r := range replies {
				reply := CommentWithUserInfo{}
				reply.CommentSchema = &replies[j]
				ownInfo, err := s.cache.GetUserBaseInfo(r.ContentOwn)
				utils.AssertErr(err, "", 500)
				reply.Own = ownInfo
				userInfo, err := s.cache.GetUserBaseInfo(r.UserID)
				utils.AssertErr(err, "", 500)
				reply.User = userInfo
				if userID != "" && r.IsDelete == false {
					_id, err := primitive.ObjectIDFromHex(userID)
					utils.AssertErr(err, "", 500)
					reply.Liked = s.cache.IsLikeComment(_id, r.ID)
				}
				comment.Reply = append(comment.Reply, reply)
			}
		} else {
			comment.Reply = []CommentWithUserInfo{}
		}
		res = append(res, comment)
	}
	return res
}

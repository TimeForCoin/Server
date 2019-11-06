package services

import (
	"github.com/TimeForCoin/Server/app/models"
	"github.com/TimeForCoin/Server/app/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageService 消息服务
type MessageService interface {
	GetSessions(userID primitive.ObjectID, page, size int64) (res []SessionListDetail)
	GetSession(userID, sessionID primitive.ObjectID, page, size int64) SessionDetail
	GetSessionByUser(userID, targetID primitive.ObjectID, page, size int64) SessionDetail
	SendSystemMessage(userID, aboutID primitive.ObjectID, title, content string) (total int64)
	SendChatMessage(userID, targetID primitive.ObjectID, msg string) primitive.ObjectID
}

// newMessageService 初始化
func newMessageService() MessageService {
	return &messageService{
		model:     models.GetModel().Message,
		userModel: models.GetModel().User,
		cache:     models.GetRedis().Cache,
		taskModel: models.GetModel().Task,
	}
}

type messageService struct {
	model     *models.MessageModel
	userModel *models.UserModel
	cache     *models.CacheModel
	taskModel *models.TaskModel
}

// SessionListDetail 会话列表信息
type SessionListDetail struct {
	*SessionDetail
	// 排除项
	Messages omit `json:"messages,omitempty"`
}

// SessionDetail 会话详细
type SessionDetail struct {
	*models.SessionSchema
	// 额外项
	Target      models.UserBaseInfo `json:"target_user"`
	UnreadCount int64               `json:"unread_count"`
	// 排除项
	User1   omit `json:"user_1,omitempty"`
	User2   omit `json:"user_2,omitempty"`
	Unread1 omit `json:"unread_1,omitempty"`
	Unread2 omit `json:"unread_2,omitempty"`
}

// makeSession 组合 Session 数据
func (s *messageService) makeSession(userID primitive.ObjectID, session models.SessionSchema) SessionDetail {
	sessionItem := SessionDetail{
		SessionSchema: &session,
	}
	if userID == session.User1 {
		sessionItem.UnreadCount = session.Unread1
	} else {
		sessionItem.UnreadCount = session.Unread2
	}
	// 用户信息
	if session.Type == models.MessageTypeChat {
		if userID == session.User1 {
			userInfo, err := s.cache.GetUserBaseInfo(session.User2)
			utils.AssertErr(err, "", 500)
			sessionItem.Target = userInfo
		} else {
			userInfo, err := s.cache.GetUserBaseInfo(session.User1)
			utils.AssertErr(err, "", 500)
			sessionItem.Target = userInfo
		}
	} else if session.Type == models.MessageTypeTask {
		// 任务信息
		taskID := session.User1
		if userID == session.User1 {
			taskID = session.User2
		}
		task, err := s.taskModel.GetTaskByID(taskID)
		utils.AssertErr(err, "", 500)
		sessionItem.Target = models.UserBaseInfo{
			ID:       taskID.Hex(),
			Nickname: task.Title,
		}
	} else if session.Type == models.MessageTypeSystem {
		sessionItem.Target = models.UserBaseInfo{
			Nickname: "系统消息",
		}
	}
	if sessionItem.Messages == nil {
		sessionItem.Messages = []models.MessageSchema{}
	}
	return sessionItem
}

// GetSessions 获取用户会话列表
func (s *messageService) GetSessions(userID primitive.ObjectID, page, size int64) (res []SessionListDetail) {
	sessions := s.model.GetSessionsByUser(userID, page, size)
	for _, session := range sessions {
		SessionDetail := s.makeSession(userID, session)
		res = append(res, SessionListDetail{
			SessionDetail: &SessionDetail,
		})
	}
	return
}

// GetSession 获取会话信息
func (s *messageService) GetSession(userID, sessionID primitive.ObjectID, page, size int64) SessionDetail {
	session, err := s.model.GetSessionWithMsgByID(sessionID, page, size)
	utils.AssertErr(err, "faked_message", 403)
	if session.User1 == userID {
		if session.Unread1 != 0 {
			utils.AssertErr(s.model.ReadMessage(sessionID, true), "", 500)
		}
	} else if session.User2 == userID {
		if session.Unread2 != 0 {
			utils.AssertErr(s.model.ReadMessage(sessionID, false), "", 500)
		}
	} else {
		utils.Assert(false, "permission_deny", 403)
	}
	return s.makeSession(userID, session)
}

// GetSessionByUser 根据用户获取会话信息
func (s *messageService) GetSessionByUser(userID, targetID primitive.ObjectID, page, size int64) SessionDetail {
	session, err := s.model.GetSessionWithMsgByUserID(userID, targetID, page, size)
	if err == nil {
		return s.makeSession(userID, session)
	}
	return s.makeSession(userID, models.SessionSchema{
		User1:    userID,
		User2:    targetID,
		Type:     models.MessageTypeChat,
		Messages: []models.MessageSchema{},
	})
}

// SendSystemMessage 发送系统消息
func (s *messageService) SendSystemMessage(userID, aboutID primitive.ObjectID, title, content string) (total int64) {
	admin, err := s.cache.GetUserBaseInfo(userID)
	utils.AssertErr(err, "faked_user", 403)
	utils.Assert(admin.Type == models.UserTypeAdmin || admin.Type == models.UserTypeRoot, "permission_deny", 403)

	users := s.userModel.GetAllUser()
	for _, userID := range users {
		_, err := s.model.AddMessage(userID, models.MessageTypeSystem, models.MessageSchema{
			Title:   title,
			Content: content,
			About:   aboutID,
		})
		utils.AssertErr(err, "", 500)
	}
	return int64(len(users))
}

// SendChatMessage 发送聊天消息
func (s *messageService) SendChatMessage(userID, targetID primitive.ObjectID, msg string) primitive.ObjectID {
	_, err := s.cache.GetUserBaseInfo(targetID)
	utils.AssertErr(err, "faked_user", 403)
	sessionID, err := s.model.AddMessage(targetID, models.MessageTypeChat, models.MessageSchema{
		UserID:  userID,
		Content: msg,
	})
	utils.AssertErr(err, "", 500)
	return sessionID
}

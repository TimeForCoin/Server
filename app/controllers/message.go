package controllers

import (
	"github.com/TimeForCoin/Server/app/libs"
	"github.com/TimeForCoin/Server/app/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageController 消息关API
type MessageController struct {
	BaseController
	Service services.MessageService
}

// BindMessageController 绑定消息路由
func BindMessageController(app *iris.Application) {
	messageService := services.GetServiceManger().Message

	messageRoute := mvc.New(app.Party("/messages"))
	messageRoute.Register(messageService, getSession().Start)
	messageRoute.Handle(new(MessageController))
}

// GetMessagesRes 获取会话列表数据
type GetMessagesRes struct {
	Pagination PaginationRes
	Data       []services.SessionListDetail
}

// Get 获取会话列表
func (c *MessageController) Get() int {
	userID := c.checkLogin()
	page, size := c.getPaginationData()

	sessions := c.Service.GetSessions(userID, page, size)
	if sessions == nil {
		sessions = []services.SessionListDetail{}
	}

	c.JSON(GetMessagesRes{
		Pagination: PaginationRes{
			Page: page,
			Size: size,
		},
		Data: sessions,
	})
	return iris.StatusOK
}

// PostMessageReq 发送消息请求
type PostMessageReq struct {
	Title   string
	Content string
	About   string
}

// Post 发送系统消息
func (c *MessageController) PostSystem() int {
	userID := c.checkLogin()
	req := PostMessageReq{}
	err := c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)
	aboutID := primitive.NilObjectID
	if req.About != "" {
		aboutID, err = primitive.ObjectIDFromHex(req.About)
		libs.AssertErr(err, "invalid_id")
	}
	c.Service.SendSystemMessage(userID, aboutID, req.Title, req.Content)

	return iris.StatusOK
}

// GetMessageRes 获取会话消息数据
type GetMessageRes struct {
	Pagination PaginationRes
	Data       services.SessionDetail
}

// GetBy 获取会话消息
func (c *MessageController) GetBy(id string) int {
	userID := c.checkLogin()
	page, size := c.getPaginationData()
	sessionID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)

	session := c.Service.GetSession(userID, sessionID, page, size)

	c.JSON(GetMessageRes{
		Pagination: PaginationRes{
			Page: page,
			Size: size,
		},
		Data: session,
	})
	return iris.StatusOK
}

// PostBy 发送消息
func (c *MessageController) PostBy(id string) int {
	userID := c.checkLogin()
	targetID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	req := PostMessageReq{}
	err = c.Ctx.ReadJSON(&req)
	libs.AssertErr(err, "invalid_value", 400)

	sessionID := c.Service.SendChatMessage(userID, targetID, req.Content)

	c.JSON(struct {
		ID string `json:"id"`
	}{
		ID: sessionID.Hex(),
	})

	return iris.StatusOK
}

// PutBy 标记会话为已读
func (c *MessageController) PutBy(id string) int {
	userID := c.checkLogin()
	sessionID, err := primitive.ObjectIDFromHex(id)
	libs.AssertErr(err, "invalid_id", 400)
	c.Service.TagRead(userID, sessionID)
	return iris.StatusOK
}

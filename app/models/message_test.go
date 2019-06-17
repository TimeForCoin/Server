package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestMessageModel(t *testing.T) {
	t.Run("InitDB", testInitDB)

	t.Run("testMessage", testMessage)

	ctx, finish := GetCtx()
	defer finish()
	err := model.Message.Collection.Drop(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Run("DisconnectDB", testDisconnectDB)
}

func testMessage(t *testing.T) {
	user1 := primitive.NewObjectID()
	user2 := primitive.NewObjectID()
	id, err := GetModel().Message.AddMessage(user2, MessageTypeChat, MessageSchema{
		UserID: user1,
		Content: "???",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(id.Hex())

	id, err = GetModel().Message.AddMessage(user1, MessageTypeChat, MessageSchema{
		UserID: user2,
		Content: "666",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(id.Hex())

	sessions := GetModel().Message.GetSessionsByUser(user1, 1, 10)
	if len(sessions) != 1 {
		t.Error(sessions)
	} else if  sessions[0].Unread1 != 1 || sessions[0].Unread2 != 1 {
		t.Error(sessions)
	}
	t.Log(sessions)

	err = GetModel().Message.ReadMessage(id, true)
	if err != nil {
		t.Error(err)
	}

	session, err := GetModel().Message.GetSessionWithMsgByID(id, 1, 10)
	if err != nil {
		t.Error(err)
	} else if len(session.Messages) != 2 {
		t.Error(session)
	} else if session.Unread1 != 0 || session.Unread2 != 1 {
		t.Error(session)
	}
	t.Log(session)

	session, err = GetModel().Message.GetSessionWithMsgByID(id, 1, 1)
	if err != nil {
		t.Error(err)
	} else if len(session.Messages) != 1 {
		t.Error(session)
	}
	t.Log(session)

	session, err = GetModel().Message.GetSessionWithMsgByID(id, 2, 5)
	if err != nil {
		t.Error(err)
	} else if len(session.Messages) != 0 {
		t.Error(session)
	}
	t.Log(session)
}
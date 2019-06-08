package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestCommentModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("testAddComment", testAddComment)

	ctx, finish := GetCtx()
	defer finish()
	err := model.Comment.Collection.Drop(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Run("DisconnectDB", testDisconnectDB)
}

func testAddComment(t *testing.T) {
	res, err := model.Comment.GetCommentsByContent(primitive.NewObjectID())
	if err != nil {
		t.Error(err)
	}
	t.Log(res)

	contentID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	contentOwn := primitive.NewObjectID()
	err = model.Comment.AddComment(contentID, contentOwn, userID, "Hello, world")
	if err != nil {
		t.Error(err)
	}

	res, err = model.Comment.GetCommentsByContent(contentID)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)

	err = model.Comment.AddComment(contentID, contentOwn, userID, "Hello, world")
	if err != nil {
		t.Error(err)
	}

	res, err = model.Comment.GetCommentsByContent(contentID)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}
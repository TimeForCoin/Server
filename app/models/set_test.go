package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestSetModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("testAddToSet", testAddToSet)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testAddToSet(t *testing.T) {
	model := GetModel().Set
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()
	res := model.GetSets(userID, SetOfLikeTask)
	if len(res.LikeTaskID) > 0 {
		t.Error()
	}
	t.Log(res)

	err := model.AddToSet(userID, taskID, SetOfLikeTask)
	if err != nil {
		t.Error(err)
	}

	err = model.AddToSet(userID, primitive.NewObjectID(), SetOfLikeTask)
	if err != nil {
		t.Error(err)
	}

	res = model.GetSets(userID, SetOfLikeTask)
	if len(res.LikeTaskID) != 2 {
		t.Error()
	}
	t.Log(res)

	err = model.AddToSet(userID, taskID, SetOfLikeTask)
	if err == nil {
		t.Error(err)
	}

	res = model.GetSets(userID, SetOfLikeTask)
	if len(res.LikeTaskID) != 2 {
		t.Error()
	}
	t.Log(res)

	err = model.AddToSet(userID, taskID, SetOfLikeComment)
	if err != nil {
		t.Error(err)
	}
	res = model.GetSets(userID, SetOfLikeComment)
	if len(res.LikeCommentID) != 1 {
		t.Error()
	}
	t.Log(res)
}
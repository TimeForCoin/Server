package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestTaskModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("InitRedis", testInitRedis)

	t.Run("testTask", testTaskModelAll)

	t.Run("DisconnectRedis", testDisconnectRedis)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testTaskModelAll(t *testing.T) {
	// 新建任务
	uid := primitive.NewObjectID().Hex()
	tid, err := model.Task.AddTask(uid)
	if err != nil {
		t.Error(err)
	}
	t.Log(tid)

	err = model.Task.SetTaskInfoByID(tid, TaskSchema{
		Title:        "req.Title",
		Type:         "run",
		Content:      "req.Content",
		Attachment:   []primitive.ObjectID{primitive.NewObjectID()},
		Location:     []string{"req.Location"},
		Tags:         []string{"req.Tags"},
		TopTime:      111111111,
		Reward:       "money",
		RewardValue:  100,
		RewardObject: "req.RewardObject",
		StartDate:    11111111,
		EndDate:      11111111,
		MaxPlayer:    100,
		MaxFinish:    100,
		AutoAccept:   true,
	})
	if err != nil {
		t.Error(err)
	}
	task, err := model.Task.GetTaskByID(tid)
	if err != nil {
		t.Error(err)
	}
	t.Log(task)
}

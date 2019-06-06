package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTaskModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("testTask", testTaskModelAll)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testTaskModelAll(t *testing.T) {
	// 新建任务
	uid := primitive.NewObjectID()
	tid, err := model.Task.AddTask(uid, TaskStatusWait)
	if err != nil {
		t.Error(err)
	}
	t.Log(tid)

	err = model.Task.SetTaskInfoByID(tid, TaskSchema{
		Title:        "req.Title",
		Type:         "run",
		Content:      "req.Content",
		Location:     []string{"req.Location"},
		Tags:         []string{"req.Tags"},
		TopTime:      111111111,
		Reward:       "money",
		RewardValue:  100,
		RewardObject: "req.RewardObject",
		StartDate:    11111111,
		EndDate:      11111111,
		MaxPlayer:    100,
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

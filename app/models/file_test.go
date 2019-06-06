package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestFileModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("testAddFile", testAddFile)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testAddFile(t *testing.T) {
	newID := primitive.NewObjectID()
	res, err := GetModel().File.GetFile(newID.Hex())
	if err == nil {
		t.Error(res)
	}
	id, err := GetModel().File.AddFile(newID, FileForTask, FileImage,
		"文件","附件", 10085,true)
	if err != nil {
		t.Error(err)
	}
	t.Log(id)

	res, err = GetModel().File.GetFile(id.Hex())
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}
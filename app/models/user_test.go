package models

import (
	"testing"
)

func TestUserModel(t *testing.T) {
	t.Run("InitDB", testInitDB)

	t.Run("AddUser", testUserModelAddUser)
	t.Run("FindUser", testUserModelFindUser)
	t.Run("UpdateUser", testUserModelUpdateUser)
	t.Run("RemoveUser", testUserModelRemoveUser)

	t.Run("DisconnectDB", testDisconnectDB)
}

func testUserModelAddUser(t *testing.T) {
	if err := model.User.AddUser("MegaShow"); err != nil {
		t.Error(err)
	}
	if err := model.User.AddUser("World"); err != nil {
		t.Error(err)
	}
}

func testUserModelFindUser(t *testing.T) {
	if res, err := model.User.FindUser("MegaShow"); err != nil {
		t.Error(err)
	} else {
		t.Log(res)
	}
	if res, err := model.User.FindUser("Hello"); err != nil {
		t.Log(err)
	} else {
		t.Log(res)
		t.Error()
	}
}

func testUserModelUpdateUser(t *testing.T) {

	// 找不到
	if err := model.User.UpdateUser("BB", "Mega"); err != nil {
		t.Log(err)
	} else {
		t.Error()
	}
	// 更新成功
	if err := model.User.UpdateUser("MegaShow", "Golang"); err != nil {
		t.Error(err)
	}
	// 找不到
	if res, err := model.User.FindUser("MegaShow"); err != nil {
		t.Log(err)
	} else {
		t.Log(res)
		t.Error()
	}
	// 找到
	if res, err := model.User.FindUser("Golang"); err != nil {
		t.Error(err)
	} else {
		t.Log(res)
	}
}

func testUserModelRemoveUser(t *testing.T) {
	if err := model.User.RemoveUser("World"); err != nil {
		t.Error(err)
	}
	if err := model.User.RemoveUser("MegaShow"); err == nil {
		t.Error()
	}
	if err := model.User.RemoveUser("Golang"); err != nil {
		t.Error(err)
	}
}

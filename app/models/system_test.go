package models

import "testing"

func TestSystemModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("testExistAutoEmail", testExistAutoEmail)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testExistAutoEmail(t *testing.T) {
	res := model.System.ExistAutoEmail("em.com")
	t.Log(res)
	err := model.System.AddAutoEmail("em.com", "大山中学")
	if err != nil {
		t.Error(err)
	}
	res = model.System.ExistAutoEmail("em.com")
	t.Log(res)
}
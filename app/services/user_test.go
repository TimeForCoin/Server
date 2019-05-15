package services

import (
	"github.com/TimeForCoin/Server/app/libs"
	"testing"
)

var userServiceTest UserService

func TestUserService(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("InitViolet", testInitViolet)
	userServiceTest = NewUserService()
	if userServiceTest == nil {
		t.Error()
	}
	t.Run("GetLoginURL", testUserServiceGetLoginURL)

	t.Run("DisconnectDB", testDisconnectDB)
}


func testUserServiceGetLoginURL(t *testing.T) {
	url, state, err := libs.GetOauth().Api.GetLoginURL("http://localhost:30233/auth")
	if err != nil {
		t.Error(err)
	}
	t.Log("url", url)
	t.Log("state", state)
}
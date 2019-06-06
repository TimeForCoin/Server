package services

import (
	"testing"

	"github.com/TimeForCoin/Server/app/libs"
)

var userServiceTest UserService

func TestUserService(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("InitViolet", testInitViolet)
	userServiceTest = newUserService()
	if userServiceTest == nil {
		t.Error()
	}
	t.Run("GetLoginURL", testUserServiceGetLoginURL)

	t.Run("DisconnectDB", testDisconnectDB)
}

func testUserServiceGetLoginURL(t *testing.T) {
	url, state, err := libs.GetOAuth().Api.GetLoginURL("http://localhost:8080/api/session/violet")
	if err != nil {
		t.Error(err)
	}
	t.Log("url", url)
	t.Log("state", state)
}

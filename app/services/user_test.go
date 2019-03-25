package services

import "testing"

var userServiceTest UserService

func TestUserService(t *testing.T) {
	t.Run("InitDB", testInitDB)
	userServiceTest = NewUserService()
	if userServiceTest == nil {
		t.Error()
	}
	t.Run("GetPong", testUserServiceGetPong)

	t.Run("DisconnectDB", testDisconnectDB)
}

func testUserServiceGetPong(t *testing.T) {
	if userServiceTest.GetPong("ping") != "pong" {
		t.Error()
	}
}

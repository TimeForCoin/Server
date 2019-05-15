package controllers

import (
	"testing"
)

func TestUserController(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("GetPing", testUserControllerGetPing)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testUserControllerGetPing(t *testing.T) {
	//e := httptest.New(t, NewApp())
	//
	//e.GET("/user/ping").Expect().Status(httptest.StatusOK).
	//	Body().Equal("pong")
}

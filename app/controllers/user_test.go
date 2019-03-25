package controllers

import (
	"github.com/kataras/iris/httptest"
	"testing"
)

func TestUserController_GetPing(t *testing.T) {
	e := httptest.New(t, NewApp())

	e.GET("/user/ping").Expect().Status(httptest.StatusOK).
		Body().Equal("pong")
}
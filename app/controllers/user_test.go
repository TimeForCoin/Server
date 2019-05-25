package controllers

import (
	"testing"
)

func TestUserController(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("DisconnectDB", testDisconnectDB)
}

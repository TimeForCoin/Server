package controllers

import (
	"testing"
)

func TestTaskController(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("DisconnectDB", testDisconnectDB)
}

package models

import (
	"testing"
)

func TestFileModel(t *testing.T) {
	t.Run("InitDB", testInitDB)

	ctx, finish := GetCtx()
	defer finish()
	err := model.File.Collection.Drop(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Run("DisconnectDB", testDisconnectDB)
}


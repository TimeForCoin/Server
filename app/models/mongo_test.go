package models

import (
	"github.com/TimeForCoin/Server/app/utils"
	"os"
	"testing"
)

func testInitDB(t *testing.T) {
	if GetModel() != nil {
		return
	}
	err := InitDB(&utils.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		t.Error(err)
	}
	if model := GetModel(); model == nil {
		t.Error()
	}
}

func testDisconnectDB(t *testing.T) {
	if err := DisconnectDB(); err != nil {
		t.Error(err)
	}
}

func TestMongo(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("DisconnectDB", testDisconnectDB)
}

package model

import (
	"github.com/TimeForCoin/Server/app/configs"
	"os"
	"testing"
)

func testInitDB(t *testing.T) {
	err := InitDB(&configs.DBConfig{
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
		User: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		t.Error(err)
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
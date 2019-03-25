package services

import (
	"os"
	"strconv"
	"testing"

	"github.com/TimeForCoin/Server/app/configs"
	"github.com/TimeForCoin/Server/app/models"
)

func testInitDB(t *testing.T) {
	err := models.InitDB(&configs.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		t.Error(err)
	}
}

func testDisconnectDB(t *testing.T) {
	if err := models.DisconnectDB(); err != nil {
		t.Error(err)
	}
}

func testInitRedis(t *testing.T) {
	DB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		t.Error(err)
	}
	err = models.InitRedis(&configs.RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       DB,
	})
	if err != nil {
		t.Error(err)
	}
}

func testDisconnectRedis(t *testing.T) {
	if err := models.DisconnectRedis(); err != nil {
		t.Error(err)
	}
}

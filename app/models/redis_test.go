package models

import (
	"os"
	"strconv"
	"testing"

	"github.com/TimeForCoin/Server/app/libs"
)

func TestRedis(t *testing.T) {
	t.Run("InitRedis", testInitRedis)
	t.Run("DisconnectRedis", testDisconnectRedis)
}

func testInitRedis(t *testing.T) {
	DB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		t.Error(err)
	}
	err = InitRedis(&libs.RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       DB,
	})
	if err != nil {
		t.Error(err)
	}
	if cache := GetCache(); cache == nil {
		t.Error()
	}
}

func testDisconnectRedis(t *testing.T) {
	if err := DisconnectRedis(); err != nil {
		t.Error(err)
	}
}

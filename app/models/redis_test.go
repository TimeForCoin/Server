package models

import (
	"github.com/TimeForCoin/Server/app/utils"
	"os"
	"strconv"
	"testing"
)

func TestRedis(t *testing.T) {
	t.Run("InitRedis", testInitRedis)
	t.Run("DisconnectRedis", testDisconnectRedis)
}

func testInitRedis(t *testing.T) {
	if GetRedis() != nil {
		return
	}
	DB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		t.Error(err)
	}
	err = InitRedis(&utils.RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       DB,
	})
	if err != nil {
		t.Error(err)
	}
	if cache := GetRedis(); cache == nil {
		t.Error()
	}
}

func testDisconnectRedis(t *testing.T) {
	if err := DisconnectRedis(); err != nil {
		t.Error(err)
	}
}

package model

import (
	"github.com/TimeForCoin/Server/app/configs"
	"os"
	"strconv"
	"testing"
)

func TestRedis(t *testing.T) {
	t.Run("InitRedis", testInitRedis)
	t.Run("DisconnectReids", testDisconnectRedis)
}

func testInitRedis(t *testing.T) {
	DB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		t.Error(err)
	}
	err = InitRedis(&configs.RedisConfig{
		Host: os.Getenv("REDIS_HOST"),
		Port: os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: DB,
	})
	if err != nil {
		t.Error(err)
	}
}

func testDisconnectRedis(t *testing.T) {
	if err := DisconnectRedis(); err != nil {
		t.Error(err)
	}
}
package model

import (
	"github.com/kataras/iris/core/errors"
	"log"

	"github.com/TimeForCoin/Server/app/configs"
	"github.com/go-redis/redis"
)

var cache *Cache

type Cache struct {
	Redis *redis.Client
}

func InitRedis(config *configs.RedisConfig) error {
	cache = &Cache{}
	cache.Redis = redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})
	pong, err := cache.Redis.Ping().Result()
	if err != nil {
		log.Println("Failure to connect Redis!!!")
		return err
	}
	if pong == "PONG" {
		log.Println("Successful connection to Redis.")
	} else {
		log.Println("Get error from Redis: " + pong)
		return errors.New("redis_error")
	}
	return nil
}

func DisconnectRedis() error {
	return cache.Redis.Close()
}

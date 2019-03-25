package model

import (
	"log"

	"github.com/kataras/iris/core/errors"

	"github.com/TimeForCoin/Server/app/configs"
	"github.com/go-redis/redis"
)

var cache *Cache

// Cache Redis 缓存
type Cache struct {
	Redis *redis.Client
}

// GetCache 获取缓存实例
func GetCache() *Cache {
	if cache == nil {
		panic("Redis isn't Initialize!")
	}
	return cache
}

// InitRedis 初始化 Redis
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

// DisconnectRedis 断开 Redis 连接
func DisconnectRedis() error {
	return cache.Redis.Close()
}

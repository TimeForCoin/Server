package models

import (
	"errors"

	"github.com/TimeForCoin/Server/app/utils"

	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
)

var redisInst *Redis

// Redis 缓存
type Redis struct {
	Client *redis.Client
	Cache  *CacheModel
}

// GetRedis 获取缓存实例
func GetRedis() *Redis {
	return redisInst
}

// InitRedis 初始化 Redis
func InitRedis(config *utils.RedisConfig) error {
	redisInst = &Redis{}
	redisInst.Client = redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})
	pong, err := redisInst.Client.Ping().Result()
	if err != nil {
		log.Error().Err(err).Msg("Failure to connect Redis!!!")
		return err
	}
	if pong != "PONG" {
		log.Error().Msg("Get error from Redis: " + pong)
		return errors.New("redis_error")
	}
	log.Info().Msg("Successful connection to Redis.")
	redisInst.Cache = &CacheModel{Redis: redisInst.Client}

	return nil
}

// DisconnectRedis 断开 Redis 连接
func DisconnectRedis() error {
	if redisInst == nil {
		return nil
	}
	err := redisInst.Client.Close()
	redisInst = nil
	return err
}

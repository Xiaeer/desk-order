package database

import (
	"context"
	"deskorder/config"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return RDB.Ping(context.Background()).Err()
}

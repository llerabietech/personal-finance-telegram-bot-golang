package db

import (
	"context"
	"personal-finance/internal/config"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis(cfg *config.Config) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic("Couldn't connect to Redis: " + err.Error())
	}

	println("✅ Connected to Redis")
}

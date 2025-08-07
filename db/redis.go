package db

import (
    "context"
    "github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis(password string) {
    RedisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",  
        Password: password,               
        DB:       0,   
    })

    // Проверка подключения
    _, err := RedisClient.Ping(Ctx).Result()
    if err != nil {
        panic("Couldn't connect to Redis: " + err.Error())
    }

    println("✅ Connected to Redis")
}
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"personal-finance/internal/config"

	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB          *sql.DB
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func InitDB(cfg *config.Config) error {
	var err error
	DB, err = sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := MigrateDB(DB, cfg); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Printf("✅ Connected to database: %s", cfg.Database.Path)
	return nil
}

func InitRedis(cfg *config.Config) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Connected to Redis: %s", cfg.Redis.Addr)
	return nil
}

func GetActiveUsersLastQuarter(cfg *config.Config) ([]int64, error) {
	rows, err := DB.Query("SELECT DISTINCT user_id FROM expenses WHERE date >= date('now', '-? month')", cfg.App.CleanupMonths)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			continue
		}
		users = append(users, chatID)
	}
	return users, nil
}

func Close() error {
	var errs []error
	
	if DB != nil {
		if err := DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}
	
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	
	return nil
}

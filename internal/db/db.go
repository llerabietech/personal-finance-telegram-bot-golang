package db

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/config"
	"personal-finance/internal/log"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB          *sql.DB
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func InitDB(cfg *config.Config) error {
	log.WithFields(logrus.Fields{
		"db_path": cfg.Database.Path,
	}).Info("Opening database connection")
	
	var err error
	DB, err = sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := MigrateDB(DB, cfg); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.WithFields(logrus.Fields{
		"db_path": cfg.Database.Path,
	}).Info("Database connection established successfully")
	return nil
}

func InitRedis(cfg *config.Config) error {
	log.WithFields(logrus.Fields{
		"redis_addr": cfg.Redis.Addr,
		"redis_db":   cfg.Redis.DB,
	}).Info("Connecting to Redis")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.WithFields(logrus.Fields{
		"redis_addr": cfg.Redis.Addr,
		"redis_db":   cfg.Redis.DB,
	}).Info("Redis connection established successfully")
	return nil
}

func GetActiveUsersLastQuarter(cfg *config.Config) ([]int64, error) {
	log.Debug("Querying active users from last quarter")
	
	rows, err := DB.Query("SELECT DISTINCT user_id FROM expenses WHERE date >= date('now', '-? month')", cfg.App.CleanupMonths)
	if err != nil {
		log.WithError(err).Error("Failed to query active users")
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			log.WithError(err).Warn("Failed to scan user ID from row")
			continue
		}
		users = append(users, chatID)
	}
	
	log.WithFields(logrus.Fields{
		"user_count": len(users),
		"months":      cfg.App.CleanupMonths,
	}).Debug("Retrieved active users from last quarter")
	
	return users, nil
}

func Close() error {
	log.Info("Closing database connections")
	var errs []error
	
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.WithError(err).Error("Failed to close database connection")
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		} else {
			log.Info("Database connection closed successfully")
		}
	}
	
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			log.WithError(err).Error("Failed to close Redis connection")
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		} else {
			log.Info("Redis connection closed successfully")
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	
	log.Info("All database connections closed successfully")
	return nil
}

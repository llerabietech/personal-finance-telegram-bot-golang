package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"personal-finance/internal/config"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
	Redis  *redis.Client
	Bot    *tgbotapi.BotAPI
}

func New(cfg *config.Config) (*App, error) {
	app := &App{
		Config: cfg,
	}

	// Initialize database
	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize Redis
	if err := app.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to init redis: %w", err)
	}

	// Initialize Telegram bot
	if err := app.initBot(); err != nil {
		return nil, fmt.Errorf("failed to init bot: %w", err)
	}

	return app, nil
}

func (a *App) initDatabase() error {
	var err error
	a.DB, err = sql.Open("sqlite3", a.Config.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := a.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("✅ Connected to database: %s", a.Config.Database.Path)
	return nil
}

func (a *App) initRedis() error {
	a.Redis = redis.NewClient(&redis.Options{
		Addr:     a.Config.Redis.Addr,
		Password: a.Config.Redis.Password,
		DB:       a.Config.Redis.DB,
	})

	ctx := context.Background()
	if _, err := a.Redis.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Printf("✅ Connected to Redis: %s", a.Config.Redis.Addr)
	return nil
}

func (a *App) initBot() error {
	bot, err := tgbotapi.NewBotAPI(a.Config.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = a.Config.Telegram.Debug
	a.Bot = bot

	log.Printf("✅ Authorized on account %s", bot.Self.UserName)
	return nil
}

func (a *App) createTables() error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY
    );
    CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        user_id INTEGER,
        limit_sum REAL,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
    CREATE TABLE IF NOT EXISTS expenses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        category_id INTEGER,
        amount REAL,
        date TEXT,
        FOREIGN KEY(category_id) REFERENCES categories(id)
    );
    CREATE TABLE IF NOT EXISTS incomes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        source TEXT,
        amount REAL,
        date TEXT
    );`

	_, err := a.DB.Exec(query)
	return err
}

func (a *App) Close() error {
	var errs []error

	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}

	return nil
}

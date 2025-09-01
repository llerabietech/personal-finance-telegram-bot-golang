package app

import (
	"database/sql"
	"fmt"
	"personal-finance/internal/config"
	"personal-finance/internal/db"
	"personal-finance/internal/i18n"
	"personal-finance/internal/log"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
	Redis  *redis.Client
	Bot    *tgbotapi.BotAPI
}

func New(cfg *config.Config) (*App, error) {
	log.Info("Initializing FinanceBot application")

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

	// Initialize translations
	if err := app.initTranslations(); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	log.Info("FinanceBot application initialized successfully")
	return app, nil
}

func (a *App) initDatabase() error {
	log.Info("Initializing database connection")
	if err := db.InitDB(a.Config); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	a.DB = db.DB
	log.Info("Database connection established")
	return nil
}

func (a *App) initRedis() error {
	log.Info("Initializing Redis connection")
	if err := db.InitRedis(a.Config); err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}
	a.Redis = db.RedisClient
	log.Info("Redis connection established")
	return nil
}

func (a *App) initBot() error {
	log.Info("Initializing Telegram bot")
	bot, err := tgbotapi.NewBotAPI(a.Config.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = a.Config.Telegram.Debug
	a.Bot = bot

	log.WithFields(logrus.Fields{
		"username": bot.Self.UserName,
		"debug":    a.Config.Telegram.Debug,
	}).Info("Telegram bot authorized successfully")
	return nil
}

func (a *App) initTranslations() error {
	log.Info("Loading translations")
	if err := i18n.LoadTranslations(); err != nil {
		return fmt.Errorf("failed to load translations: %w", err)
	}
	log.Info("Translations loaded successfully")
	return nil
}

func (a *App) Close() error {
	log.Info("Shutting down FinanceBot application")
	return db.Close()
}

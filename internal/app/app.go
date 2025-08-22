package app

import (
	"database/sql"
	"fmt"
	"log"
	"personal-finance/internal/config"
	"personal-finance/internal/db"
	"personal-finance/internal/i18n"

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

	// Initialize translations
	if err := i18n.LoadTranslations(); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	return app, nil
}

func (a *App) initDatabase() error {
	if err := db.InitDB(a.Config); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	a.DB = db.DB
	return nil
}

func (a *App) initRedis() error {
	if err := db.InitRedis(a.Config); err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}
	a.Redis = db.RedisClient
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



func (a *App) Close() error {
	return db.Close()
}

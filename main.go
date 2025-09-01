package main

import (
	"personal-finance/internal/app"
	"personal-finance/internal/bot"
	"personal-finance/internal/config"
	"personal-finance/internal/log"
	"personal-finance/internal/scheduler"
)

func main() {
	// Initialize logging first (will use defaults until config is loaded)
	log.Init("info", "text")
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Reinitialize logging with config
	log.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info("Starting FinanceBot application")

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer application.Close()

	log.Info("Application initialized successfully, starting services")
	
	scheduler.StartScheduler(application.Bot, application.DB, application.Redis, cfg)
	bot.StartBot(application.Bot, application.DB, application.Redis, cfg)
}

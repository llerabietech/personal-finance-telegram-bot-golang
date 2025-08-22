package main

import (
	"log"
	"personal-finance/internal/app"
	"personal-finance/internal/bot"
	"personal-finance/internal/config"
	"personal-finance/internal/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer application.Close()

	scheduler.StartScheduler(application.Bot, application.DB, application.Redis, cfg)
	bot.StartBot(application.Bot, application.DB, application.Redis, cfg)
}

package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"personal-finance/internal/config"
	"personal-finance/internal/repository"
	"personal-finance/internal/service"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMonthlyReport(bot *tgbotapi.BotAPI, ctx context.Context, db *sql.DB, redis *redis.Client, cfg *config.Config) {
	users, err := repository.GetActiveUsersLastQuarter(ctx, db, cfg)
	if err != nil {
		fmt.Println("Error receiving users:", err)
		return
	}

	for _, chatID := range users {
		report, err := service.GenerateReport(ctx, db, redis, chatID, cfg)
		if err != nil {
			log.Fatalf("Error generating report: %v", err)
			continue
		}
		if report == "" {
			log.Fatalf("No data for this month")
			continue
		}
		msg := tgbotapi.NewMessage(chatID, report)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

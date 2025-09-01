package scheduler

import (
	"context"
	"database/sql"
	"personal-finance/internal/config"
	"personal-finance/internal/handler"
	"time"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// restarts the background check for the monthly report
func StartScheduler(bot *tgbotapi.BotAPI, db *sql.DB, redis *redis.Client, cfg *config.Config) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			now := time.Now()
			if now.Day() == 1 && now.Hour() == cfg.App.ReportHour && now.Minute() < cfg.App.ReportMinute+10 {
				ctx := context.Background()
				handler.HandleMonthlyReport(bot, ctx, db, redis, cfg)
			}

			<-ticker.C
		}
	}()
}

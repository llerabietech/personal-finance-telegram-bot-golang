package bot

import (
    "time"
    "personal-finance/commands"
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// restarts the background check for the monthly report
func StartScheduler(bot *tgbotapi.BotAPI) {
    go func() {
        ticker := time.NewTicker(24 * time.Hour)
        defer ticker.Stop()

        for {
            now := time.Now()
            if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 10 {
                commands.SendMonthlyReport(bot)
            }

            <-ticker.C
        }
    }()
}
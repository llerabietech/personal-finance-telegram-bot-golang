package bot

import (
    "time"
    "personal-finance/commands"
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartScheduler запускает фоновую проверку для ежемесячного отчёта
func StartScheduler(bot *tgbotapi.BotAPI) {
    go func() {
        ticker := time.NewTicker(24 * time.Hour) // Проверяем раз в день
        defer ticker.Stop()

        for {
            now := time.Now()
            // Если сегодня 1-е число и прошло 00:00 — отправляем отчёт за прошлый месяц
            if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 10 {
                commands.SendMonthlyReport(bot)
            }

            // Ждём следующий тик
            <-ticker.C
        }
    }()
}
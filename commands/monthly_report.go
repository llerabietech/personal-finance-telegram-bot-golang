package commands

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"personal-finance/db"
	"strings"
	"time"
)

// SendMonthlyReport отправляет отчёт за предыдущий месяц всем пользователям
func SendMonthlyReport(bot *tgbotapi.BotAPI) {
	now := time.Now()
	// Предыдущий месяц: например, если сейчас 1 апреля — берём март
	prevMonth := now.AddDate(0, -1, 0)
	monthStr := prevMonth.Format("2006-01")

	users, err := db.GetAllUsers()
	if err != nil {
		fmt.Println("Ошибка получения пользователей:", err)
		return
	}

	for _, chatID := range users {
		report := generateReportForUser(chatID, monthStr, prevMonth)
		if report != "" {
			msg := tgbotapi.NewMessage(chatID, report)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	}
}

// generateReportForUser — генерирует отчёт для одного пользователя
func generateReportForUser(chatID int64, monthStr string, month time.Time) string {
	rows, err := db.DB.Query(`
        SELECT 
            c.name,
            SUM(e.amount),
            c.limit_sum
        FROM expenses e
        JOIN categories c ON e.category_id = c.id
        WHERE e.user_id = ? AND e.date LIKE ?
        GROUP BY c.name, c.limit_sum`, chatID, monthStr+"%")
	if err != nil {
		return ""
	}
	defer rows.Close()

	var lines []string
	var totalSpent float64
	var overLimit int

	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		totalSpent += spent

		status := "✅"
		if spent > limit {
			status = "❌"
			overLimit++
		}

		lines = append(lines, fmt.Sprintf("- %s: %.2f ₽ / %.2f ₽ %s", strings.Title(name), spent, limit, status))
	}

	if len(lines) == 0 {
		return ""
	}

	monthName := month.Format("January")
	emoji := "🟢"
	if overLimit > 0 {
		emoji = "🟡"
	}
	if overLimit > 2 {
		emoji = "🔴"
	}

	return fmt.Sprintf(`%s **Месячный отчёт за %s** %s

💸 *Общие траты*: %.2f ₽

📊 *По категориям*:
%s

📌 *Превышено лимитов*: %d шт.

Спасибо, что используете финансового помощника! 💼`,
		emoji, strings.Title(strings.ToLower(monthName)), emoji, totalSpent,
		strings.Join(lines, "\n"), overLimit)
}

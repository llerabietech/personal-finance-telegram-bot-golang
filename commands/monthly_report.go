package commands

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"personal-finance/db"
	"strings"
	"time"
)

// отправляет отчёт за предыдущий месяц всем пользователям
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

	// После отправки всех отчётов — чистим старые траты
	CleanupOldExpenses()
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

	var totalIncome float64
	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = ? AND date LIKE ?",
		chatID, monthStr+"%").Scan(&totalIncome)

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
	balance := totalIncome - totalSpent
	emoji := "🟢"
	if balance < 0 {
		emoji = "🔴"
	} else if balance < totalIncome*0.1 {
		emoji = "🟡"
	}

	return fmt.Sprintf(`%s **Месячный отчёт за %s** %s

💼 *Доходы*: %.2f ₽
💸 *Расходы*: %.2f ₽
📊 *Баланс*: %.2f ₽

...
`, emoji, monthName, emoji, totalIncome, totalSpent, balance)
}

// удаляет траты старше 3 месяцев
func CleanupOldExpenses() {
	// Определяем дату: всё, что раньше 3 месяцев — удаляем
	threeMonthsAgo := time.Now().AddDate(0, -3, 0).Format("2006-01-02")

	result, err := db.DB.Exec("DELETE FROM expenses WHERE date < ?", threeMonthsAgo)
	if err != nil {
		println("Ошибка при удалении старых трат:", err.Error())
		return
	}

	rows, _ := result.RowsAffected()
	println("Очистка: удалено", rows, "старых трат (до", threeMonthsAgo+")")
}

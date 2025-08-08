package commands

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"personal-finance/db"
	"personal-finance/i18n"
	"personal-finance/state"
	"personal-finance/utils"
	"strings"
	"time"
)

// sends a report for the previous month to all users
func SendMonthlyReport(bot *tgbotapi.BotAPI) {
	now := time.Now()
	prevMonth := now.AddDate(0, -1, 0)
	monthStr := prevMonth.Format("2006-01")

	users, err := db.GetActiveUsersLastQuarter()
	if err != nil {
		fmt.Println("Error receiving users:", err)
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

	CleanupOldExpenses()
}

// generates a report for a single user
func generateReportForUser(chatID int64, monthStr string, month time.Time) string {
	lang, err := state.GetUserLanguage(chatID)
	if err != nil {
		lang = "en"
	}

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
	var totalSpent, totalIncome, balance float64
	var overLimit int

	db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = ? AND date LIKE ?",
		chatID, monthStr+"%").Scan(&totalIncome)

	totalSpent = 0
	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		totalSpent += spent

		status := i18n.T("monthly_status_under", lang)
		if spent > limit {
			status = i18n.T("monthly_status_over", lang)
			overLimit++
		}

		displayName := utils.Title.String(name)
		lines = append(lines, fmt.Sprintf("- %s: %.2f ₽ / %.2f ₽ %s", displayName, spent, limit, status))
	}

	if len(lines) == 0 && totalIncome == 0 {
		return ""
	}

	balance = totalIncome - totalSpent
	emoji := "🟢"
	if balance < 0 {
		emoji = "🔴"
	} else if balance < totalIncome*0.1 {
		emoji = "🟡"
	}

	monthName := getMonthName(month, lang)

	report := utils.FormatAmount(fmt.Sprintf(`%s `+i18n.T("monthly_report_title", lang)+` %s

`+i18n.T("monthly_income", lang)+`
`+i18n.T("monthly_expenses", lang)+`
`+i18n.T("monthly_balance", lang)+`

`+i18n.T("monthly_categories", lang)+`:
%s

`+i18n.T("monthly_over_limits", lang)+`

`+i18n.T("monthly_thanks", lang)+`
`,
		emoji, monthName, emoji,
		totalIncome, totalSpent, balance,
		strings.Join(lines, "\n"),
		overLimit,
	), lang)

	return report
}

// TODO helper
func getMonthName(t time.Time, lang string) string {
	months := map[string][]string{
		"ru": {"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
			"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"},
		"en": {"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"},
	}
	if m, ok := months[lang]; ok {
		return m[t.Month()-1]
	}
	return t.Month().String()
}

func CleanupOldExpenses() {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0).Format("2006-01-02")

	result, err := db.DB.Exec("DELETE FROM expenses WHERE date < ?", threeMonthsAgo)
	if err != nil {
		println("Error when deleting old expenses:", err.Error())
		return
	}

	rows, _ := result.RowsAffected()
	println("Cleanup: deleted", rows, "old expenses (before", threeMonthsAgo+")")
}

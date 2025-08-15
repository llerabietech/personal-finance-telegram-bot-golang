package commands

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/i18n"
	"personal-finance/internal/config"
	"personal-finance/state"
	"personal-finance/utils"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// sends a report for the previous month to all users
func SendMonthlyReport(bot *tgbotapi.BotAPI, ctx context.Context, db *sql.DB, redis *redis.Client, cfg *config.Config) {
	now := time.Now()
	prevMonth := now.AddDate(0, -1, 0)
	monthStr := prevMonth.Format(cfg.App.MonthFormat)

	users, err := GetActiveUsersLastQuarter(ctx, db, cfg)
	if err != nil {
		fmt.Println("Error receiving users:", err)
		return
	}

	for _, chatID := range users {
		report := generateReportForUser(ctx, db, redis, chatID, monthStr, prevMonth, cfg)
		if report != "" {
			msg := tgbotapi.NewMessage(chatID, report)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	}

	CleanupOldExpenses(ctx, db, cfg)
}

// generates a report for a single user
func generateReportForUser(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, monthStr string, month time.Time, cfg *config.Config) string {
	lang, err := state.GetUserLanguage(ctx, redis, chatID)
	if err != nil {
		lang = "en"
	}

	rows, err := db.QueryContext(ctx, `
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

	db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = ? AND date LIKE ?",
		chatID, monthStr+"%").Scan(&totalIncome)

	totalSpent = 0
	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		totalSpent += spent

		status := i18n.T("monthly_status_under", lang, cfg)
		if spent > limit {
			status = i18n.T("monthly_status_over", lang, cfg)
			overLimit++
		}

		displayName := utils.Title.String(name)
		lines = append(lines, fmt.Sprintf("- %s: %.2f %s / %.2f %s %s", displayName, spent, cfg.App.CurrencySymbol, limit, cfg.App.CurrencySymbol, status))
	}

	if len(lines) == 0 && totalIncome == 0 {
		return ""
	}

	balance = totalIncome - totalSpent
	emoji := cfg.App.StatusEmojis.BalanceGood
	if balance < 0 {
		emoji = cfg.App.StatusEmojis.BalanceBad
	} else if balance < totalIncome*(cfg.App.BalanceWarningThreshold/100) {
		emoji = cfg.App.StatusEmojis.BalanceWarning
	}

	monthName := utils.GetMonthName(month, lang, cfg)

	report := utils.FormatAmount(fmt.Sprintf(`%s `+i18n.T("monthly_report_title", lang, cfg)+` %s

`+i18n.T("monthly_income", lang, cfg)+`
`+i18n.T("monthly_expenses", lang, cfg)+`
`+i18n.T("monthly_balance", lang, cfg)+`

`+i18n.T("monthly_categories", lang, cfg)+`:
%s

`+i18n.T("monthly_over_limits", lang, cfg)+`

`+i18n.T("monthly_thanks", lang, cfg)+`
`,
		emoji, monthName, emoji,
		totalIncome, totalSpent, balance,
		strings.Join(lines, "\n"),
		overLimit,
	), lang, cfg)

	return report
}

func CleanupOldExpenses(ctx context.Context, db *sql.DB, cfg *config.Config) {
	threeMonthsAgo := time.Now().AddDate(0, -cfg.App.CleanupMonths, 0).Format(cfg.App.DateFormat)

	result, err := db.ExecContext(ctx, "DELETE FROM expenses WHERE date < ?", threeMonthsAgo)
	if err != nil {
		println("Error when deleting old expenses:", err.Error())
		return
	}

	rows, _ := result.RowsAffected()
	println("Cleanup: deleted", rows, "old expenses (before", threeMonthsAgo+")")
}

// GetActiveUsersLastQuarter moved from db package
func GetActiveUsersLastQuarter(ctx context.Context, db *sql.DB, cfg *config.Config) ([]int64, error) {
	rows, err := db.QueryContext(ctx, "SELECT DISTINCT user_id FROM expenses WHERE date >= date('now', '-? month')", cfg.App.CleanupMonths)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			continue
		}
		users = append(users, chatID)
	}
	return users, nil
}

package commands

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/i18n"
	"personal-finance/internal/config"
	"personal-finance/state"
	"personal-finance/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)




func ListLimits(ctx context.Context, db *sql.DB, chatID int64, lang string, cfg *config.Config) string {
	rows, err := db.QueryContext(ctx, "SELECT name, limit_sum FROM categories WHERE user_id = ?", chatID)
	if err != nil {
		return i18n.T("error_load_limits", lang, cfg)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		var limit float64
		rows.Scan(&name, &limit)
		result = append(result, fmt.Sprintf("%s: %.2f %s", name, limit, cfg.App.CurrencySymbol))
	}

	return i18n.T("limits2", lang, cfg) + strings.Join(result, "\n")
}

func GetAnalytics(ctx context.Context, db *sql.DB, chatID int64, lang string, cfg *config.Config) string {
	month := time.Now().Format(cfg.App.MonthFormat)

	var totalIncome float64
	err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = ? AND date LIKE ?",
		chatID, month+"%").Scan(&totalIncome)
	if err != nil {
		totalIncome = 0
	}

	var totalExpenses float64
	err = db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM expenses e JOIN categories c ON e.category_id = c.id WHERE e.user_id = ? AND e.date LIKE ?",
		chatID, month+"%").Scan(&totalExpenses)
	if err != nil {
		totalExpenses = 0
	}

	rows, err := db.QueryContext(ctx, `
        SELECT c.name, SUM(e.amount), c.limit_sum 
        FROM expenses e
        JOIN categories c ON e.category_id = c.id
        WHERE e.user_id = ? AND e.date LIKE ?
        GROUP BY c.name, c.limit_sum`, chatID, month+"%")
	if err != nil {
		return i18n.T("limits2", lang, cfg)
	}
	defer rows.Close()

	var report []string
	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		status := cfg.App.StatusEmojis.Success
		if spent > limit {
			status = cfg.App.StatusEmojis.Error
		}
		report = append(report, fmt.Sprintf("• %s: %.2f %s / %.2f %s %s", name, spent, cfg.App.CurrencySymbol, limit, cfg.App.CurrencySymbol, status))
	}

	balance := totalIncome - totalExpenses
	balanceEmoji := cfg.App.StatusEmojis.BalanceGood
	if balance < 0 {
		balanceEmoji = cfg.App.StatusEmojis.BalanceBad
	} else if balance < totalIncome*(cfg.App.BalanceWarningThreshold/100) {
		balanceEmoji = cfg.App.StatusEmojis.BalanceWarning
	}

	details := "—"
	if len(report) > 0 {
		details = strings.Join(report, "\n")
	}

	return utils.FormatAmount(i18n.Tf("analytics_title", lang, cfg, utils.GetMonthName(time.Now(), lang, cfg), totalIncome, totalExpenses, balanceEmoji, balance, details), lang, cfg)
}



func UpdateLimit(ctx context.Context, db *sql.DB, chatID int64, categoryName string, newLimit float64, lang string, cfg *config.Config) string {
	name := strings.ToLower(strings.TrimSpace(categoryName))
	if name == "" {
		return i18n.T("error_category_name_is_empty", lang, cfg)
	}

	var categoryID int
	var currentLimit float64

	err := db.QueryRowContext(ctx, `
        SELECT id, limit_sum 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, name).Scan(&categoryID, &currentLimit)

	if err == sql.ErrNoRows {
		return i18n.Tf("category_not_found", lang, cfg, utils.Title.String(name))
	} else if err != nil {
		return i18n.T("error_found_category", lang, cfg) + err.Error()
	}

	_, err = db.ExecContext(ctx, "UPDATE categories SET limit_sum = ? WHERE id = ?", newLimit, categoryID)
	if err != nil {
		return i18n.T("error_update_limit", lang, cfg)
	}

	return utils.FormatAmount(i18n.Tf("updated_limit", lang, cfg, utils.Title.String(name), currentLimit, newLimit), lang, cfg)
}



func ConfirmDelete(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, answer string, lang string, cfg *config.Config) string {
	categoryIDStr, err := state.GetTempData(ctx, redis, chatID)
	if err != nil {
		state.Clear(ctx, redis, chatID)
		return i18n.T("error_old_data", lang, cfg)
	}

	// Проверяем, является ли ответ одним из подтверждающих слов
	isConfirmed := false
	for _, word := range cfg.App.ConfirmationWords {
		if strings.EqualFold(answer, word){
			isConfirmed = true
			break
		}
	}
	if !isConfirmed {
		state.Clear(ctx, redis, chatID)
		return i18n.T("delete_cancelled", lang, cfg)
	}

	var categoryID int
	fmt.Sscanf(categoryIDStr, "%d", &categoryID)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return i18n.T("error_delete", lang, cfg)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM expenses WHERE category_id = ?", categoryID)
	if err != nil {
		tx.Rollback()
		return i18n.T("error_delete_expense", lang, cfg)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", categoryID)
	if err != nil {
		tx.Rollback()
		return i18n.T("error_delete_category", lang, cfg)
	}

	err = tx.Commit()
	if err != nil {
		return i18n.T("error_confirm_changes", lang, cfg)
	}

	state.Clear(ctx, redis, chatID)
	return i18n.T("category_deleted", lang, cfg)
}

func AddIncome(ctx context.Context, db *sql.DB, chatID int64, input string, lang string, cfg *config.Config) string {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return i18n.T("invalid_format2", lang, cfg)
	}

	source := strings.ToLower(strings.TrimSpace(parts[0]))
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		return i18n.T("error_summ", lang, cfg)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO incomes (user_id, source, amount, date) VALUES (?, ?, ?, ?)",
		chatID, source, amount, time.Now().Format(cfg.App.DateFormat))
	if err != nil {
		return i18n.T("error_save_income", lang, cfg)
	}

	return utils.FormatAmount(i18n.Tf("add_income", lang, cfg, utils.Title.String(source), amount), lang, cfg)
}

func IsPotentialIncome(ctx context.Context, db *sql.DB, chatID int64, text string) bool {
	parts := strings.Fields(text)
	if len(parts) != 2 {
		return false
	}
	_, err := strconv.ParseFloat(parts[1], 64)
	return err == nil
}

package commands

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"personal-finance/internal/i18n"
	"personal-finance/internal/config"
	"personal-finance/state"
	"personal-finance/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetLanguageKeyboard(cfg *config.Config) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🇷🇺 Русский"),
			tgbotapi.NewKeyboardButton("🇬🇧 English"),
		),
	)
}

func GetMainMenu(lang string, cfg *config.Config) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("expenses", lang, cfg)),
			tgbotapi.NewKeyboardButton(i18n.T("income", lang, cfg)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("categories", lang, cfg)),
			tgbotapi.NewKeyboardButton(i18n.T("limits", lang, cfg)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("analytics", lang, cfg)),
		),
	)
}

func GetCategoriesMenu(lang string, cfg *config.Config) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("list_categories", lang, cfg)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("add_category", lang, cfg)),
			tgbotapi.NewKeyboardButton(i18n.T("delete_category", lang, cfg)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("back", lang, cfg)),
		),
	)
}

func GetLimitsMenu(lang string, cfg *config.Config) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("limits_list", lang, cfg)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("change_limit", lang, cfg)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("back", lang, cfg)),
		),
	)
}

func AddCategory(ctx context.Context, db *sql.DB, chatID int64, input string, lang string, cfg *config.Config) string {
	parts := strings.Fields(input)

	name := strings.ToLower(parts[1])
	limit, err := strconv.ParseFloat(parts[2], 64)
	if err != nil || limit <= 0 {
		return i18n.T("correct_digit", lang, cfg)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)
	if err != nil {
		return i18n.T("error_add_category", lang, cfg)
	}

	return utils.FormatAmount(i18n.Tf("category_created", lang, cfg, name, limit), lang, cfg)
}

func AddExpense(bot *tgbotapi.BotAPI, ctx context.Context, db *sql.DB, chatID int64, input string, lang string, cfg *config.Config) string {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return i18n.T("invalid_format_expense", lang, cfg)
	}

	categoryInput := strings.ToLower(strings.TrimSpace(parts[0]))
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return i18n.T("invalid_amount", lang, cfg)
	}

	var categoryID int
	var categoryName string
	err = db.QueryRowContext(ctx, `
        SELECT id, name 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, categoryInput).Scan(&categoryID, &categoryName)

	if err == sql.ErrNoRows {
		return i18n.Tf("category_not_found", lang, cfg, categoryInput)
	} else if err != nil {
		return i18n.T("error_found_category", lang, cfg)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO expenses (user_id, category_id, amount, date) VALUES (?, ?, ?, ?)",
		chatID, categoryID, amount, time.Now().Format(cfg.App.DateFormat))
	if err != nil {
		return i18n.T("error_save_expense", lang, cfg)
	}

	go CheckLimitAndNotify(bot, ctx, db, chatID, categoryID, categoryName, lang, cfg)

	return utils.FormatAmount(i18n.Tf("add_expense", lang, cfg, utils.Title.String(categoryName), amount), lang, cfg)
}

func ListCategories(ctx context.Context, db *sql.DB, chatID int64, lang string, cfg *config.Config) string {
	rows, err := db.QueryContext(ctx, "SELECT name FROM categories WHERE user_id = ?", chatID)
	if err != nil {
		return i18n.T("error_load_category", lang, cfg)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}

	if len(categories) == 0 {
		return i18n.T("error_empty_category", lang, cfg)
	}
	return i18n.T("categories2", lang, cfg) + strings.Join(categories, "\n• ")
}

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

func IsPotentialExpense(ctx context.Context, db *sql.DB, chatID int64, text string) bool {
	parts := strings.Fields(text)
	if len(parts) != 2 {
		return false
	}

	categoryName := strings.ToLower(strings.TrimSpace(parts[0]))
	amountStr := parts[1]

	_, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return false
	}

	var count int
	err = db.QueryRowContext(ctx, `
        SELECT COUNT(*) 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, categoryName).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func CreateCategory(ctx context.Context, db *sql.DB, chatID int64, name string, limit float64, lang string, cfg *config.Config) string {
	_, err := db.ExecContext(ctx, "INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)
	if err != nil {
		return i18n.T("error_create_category", lang, cfg)
	}
	return utils.FormatAmount(i18n.Tf("category_created", lang, cfg, utils.Title.String(name), limit), lang, cfg)
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

func HandleNewCategoryName(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, text string, lang string, cfg *config.Config) string {
	name := strings.ToLower(strings.TrimSpace(text))
	if name == "" {
		return i18n.T("empty_name", lang, cfg)
	}

	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories WHERE user_id = ? AND LOWER(name) = ?", chatID, name).Scan(&count)
	if err != nil {
		return err.Error()
	}
	if count > 0 {
		state.Clear(ctx, redis, chatID)
		return i18n.T("error_category_already_exist", lang, cfg)
	}

	state.SetTempData(ctx, redis, chatID, name)
	state.SetState(ctx, redis, chatID, state.AwaitingCategoryLimit)

	return i18n.Tf("enter_limit2", lang, cfg, utils.Title.String(name))
}

func CheckLimitAndNotify(bot *tgbotapi.BotAPI, ctx context.Context, db *sql.DB, chatID int64, categoryID int, categoryName string, lang string, cfg *config.Config) {
	month := time.Now().Format(cfg.App.MonthFormat)

	var spent float64
	err := db.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(amount), 0)
        FROM expenses
        WHERE user_id = ? AND category_id = ? AND date LIKE ?`,
		chatID, categoryID, month+"%").Scan(&spent)
	if err != nil {
		return
	}

	var limitSum float64
	err = db.QueryRowContext(ctx, "SELECT limit_sum FROM categories WHERE id = ? AND user_id = ?",
		categoryID, chatID).Scan(&limitSum)
	if err != nil || limitSum <= 0 {
		return
	}

	percent := (spent / limitSum) * 100

	var msgText string
	sent := false

	if percent >= cfg.App.LimitOverloadThreshold {
		msgText = utils.FormatAmount(i18n.Tf("limit_is_overloaded", lang, cfg, utils.Title.String(categoryName), spent, limitSum), lang, cfg)
		sent = true
	} else if percent >= cfg.App.LimitWarningThreshold {
		msgText = utils.FormatAmount(i18n.Tf("limit_warning", lang, cfg, math.Round(percent), utils.Title.String(categoryName), spent, limitSum), lang, cfg)
		sent = true
	}

	if sent {
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

func HandleDeleteCategory(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, categoryName string, lang string, cfg *config.Config) string {
	name := strings.ToLower(strings.TrimSpace(categoryName))

	var categoryID int
	var limit float64
	err := db.QueryRowContext(ctx, "SELECT id, limit_sum FROM categories WHERE user_id = ? AND LOWER(name) = ?",
		chatID, name).Scan(&categoryID, &limit)

	if err == sql.ErrNoRows {
		return i18n.T("category_not_found2", lang, cfg)
	} else if err != nil {
		return i18n.T("error_found_category", lang, cfg)
	}

	state.SetTempData(ctx, redis, chatID, fmt.Sprintf("%d", categoryID))
	state.SetState(ctx, redis, chatID, state.ConfirmDeleteCategory)

	displayName := utils.Title.String(name)
	return utils.FormatAmount(i18n.Tf("confirm_delete", lang, cfg, displayName, limit), lang, cfg)
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

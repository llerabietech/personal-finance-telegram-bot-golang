package commands

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"personal-finance/db"
	"personal-finance/i18n"
	"personal-finance/state"
	"personal-finance/utils"
	"strconv"
	"strings"
	"time"
)

func GetLanguageKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🇷🇺 Русский"),
			tgbotapi.NewKeyboardButton("🇬🇧 English"),
		),
	)
}

func GetMainMenu(lang string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("expenses", lang)),
			tgbotapi.NewKeyboardButton(i18n.T("income", lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("categories", lang)),
			tgbotapi.NewKeyboardButton(i18n.T("limits", lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("analytics", lang)),
		),
	)
}

func GetCategoriesMenu(lang string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("list_categories", lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("add_category", lang)),
			tgbotapi.NewKeyboardButton(i18n.T("delete_category", lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("back", lang)),
		),
	)
}

func GetLimitsMenu(lang string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("limits_list", lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("change_limit", lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.T("back", lang)),
		),
	)
}

func AddCategory(chatID int64, input string, lang string) string {
	parts := strings.Fields(input)

	name := strings.ToLower(parts[1])
	limit, err := strconv.ParseFloat(parts[2], 64)
	if err != nil || limit <= 0 {
		return i18n.T("correct_digit", lang)
	}

	_, err = db.DB.Exec("INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)
	if err != nil {
		return i18n.T("error_add_category", lang)
	}

	return utils.FormatAmount(i18n.Tf("category_created", lang, name, limit), lang)
}

func AddExpense(bot *tgbotapi.BotAPI, chatID int64, input string, lang string) string {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return i18n.T("invalid_format_expense", lang)
	}

	categoryInput := strings.ToLower(strings.TrimSpace(parts[0]))
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return i18n.T("invalid_amount", lang)
	}

	var categoryID int
	var categoryName string
	err = db.DB.QueryRow(`
        SELECT id, name 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, categoryInput).Scan(&categoryID, &categoryName)

	if err == sql.ErrNoRows {
		return i18n.Tf("category_not_found", lang, categoryInput)
	} else if err != nil {
		return i18n.T("error_found_category", lang)
	}

	_, err = db.DB.Exec("INSERT INTO expenses (user_id, category_id, amount, date) VALUES (?, ?, ?, ?)",
		chatID, categoryID, amount, time.Now().Format("2006-01-02"))
	if err != nil {
		return i18n.T("error_save_expense", lang)
	}

	go CheckLimitAndNotify(bot, chatID, categoryID, categoryName, lang)

	return utils.FormatAmount(i18n.Tf("add_expense", lang, utils.Title.String(categoryName), amount), lang)
}

func ListCategories(chatID int64, lang string) string {
	rows, err := db.DB.Query("SELECT name FROM categories WHERE user_id = ?", chatID)
	if err != nil {
		return i18n.T("error_load_category", lang)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}

	if len(categories) == 0 {
		return i18n.T("error_empty_category", lang)
	}
	return i18n.T("categories2", lang) + strings.Join(categories, "\n• ")
}

func ListLimits(chatID int64, lang string) string {
	rows, err := db.DB.Query("SELECT name, limit_sum FROM categories WHERE user_id = ?", chatID)
	if err != nil {
		return i18n.T("error_load_limits", lang)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		var limit float64
		rows.Scan(&name, &limit)
		result = append(result, fmt.Sprintf("%s: %.2f ₽", name, limit))
	}

	return i18n.T("limits2", lang) + strings.Join(result, "\n")
}

func GetAnalytics(chatID int64, lang string) string {
	month := time.Now().Format("2006-01")

	var totalIncome float64
	err := db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = ? AND date LIKE ?",
		chatID, month+"%").Scan(&totalIncome)
	if err != nil {
		totalIncome = 0
	}

	var totalExpenses float64
	err = db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses e JOIN categories c ON e.category_id = c.id WHERE e.user_id = ? AND e.date LIKE ?",
		chatID, month+"%").Scan(&totalExpenses)
	if err != nil {
		totalExpenses = 0
	}

	rows, err := db.DB.Query(`
        SELECT c.name, SUM(e.amount), c.limit_sum 
        FROM expenses e
        JOIN categories c ON e.category_id = c.id
        WHERE e.user_id = ? AND e.date LIKE ?
        GROUP BY c.name, c.limit_sum`, chatID, month+"%")
	if err != nil {
		return i18n.T("limits2", lang)
	}
	defer rows.Close()

	var report []string
	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		status := "✅"
		if spent > limit {
			status = "❌"
		}
		report = append(report, fmt.Sprintf("• %s: %.2f ₽ / %.2f ₽ %s", name, spent, limit, status))
	}

	balance := totalIncome - totalExpenses
	balanceEmoji := "🟢"
	if balance < 0 {
		balanceEmoji = "🔴"
	} else if balance < totalIncome*0.1 {
		balanceEmoji = "🟡"
	}

	details := "—"
	if len(report) > 0 {
		details = strings.Join(report, "\n")
	}

	return utils.FormatAmount(i18n.Tf("analytics_title", lang, utils.GetMonthName(time.Now(), lang), totalIncome, totalExpenses, balanceEmoji, balance, details), lang)
}

func IsPotentialExpense(chatID int64, text string) bool {
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
	err = db.DB.QueryRow(`
        SELECT COUNT(*) 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, categoryName).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func CreateCategory(chatID int64, name string, limit float64, lang string) string {
	_, err := db.DB.Exec("INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)
	if err != nil {
		return i18n.T("error_create_category", lang)
	}
	return utils.FormatAmount(i18n.Tf("category_created", lang, utils.Title.String(name), limit), lang)
}

func UpdateLimit(chatID int64, categoryName string, newLimit float64, lang string) string {
	name := strings.ToLower(strings.TrimSpace(categoryName))
	if name == "" {
		return i18n.T("error_category_name_is_empty", lang)
	}

	var categoryID int
	var currentLimit float64

	err := db.DB.QueryRow(`
        SELECT id, limit_sum 
        FROM categories 
        WHERE user_id = ? AND LOWER(name) = ?`,
		chatID, name).Scan(&categoryID, &currentLimit)

	if err == sql.ErrNoRows {
		return i18n.Tf("category_not_found", lang, utils.Title.String(name))
	} else if err != nil {
		return i18n.T("error_found_category", lang) + err.Error()
	}

	_, err = db.DB.Exec("UPDATE categories SET limit_sum = ? WHERE id = ?", newLimit, categoryID)
	if err != nil {
		return i18n.T("error_update_limit", lang)
	}

	return utils.FormatAmount(i18n.Tf("updated_limit", lang, utils.Title.String(name), currentLimit, newLimit), lang)
}

func HandleNewCategoryName(chatID int64, text string, lang string) string {
	name := strings.ToLower(strings.TrimSpace(text))
	if name == "" {
		return i18n.T("empty_name", lang)
	}

	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE user_id = ? AND LOWER(name) = ?", chatID, name).Scan(&count)
	if err != nil {
		return err.Error()
	}
	if count > 0 {
		state.Clear(chatID)
		return i18n.T("error_category_already_exist", lang)
	}

	state.SetTempData(chatID, name)
	state.SetState(chatID, state.AwaitingCategoryLimit)

	return i18n.Tf("enter_limit2", lang, utils.Title.String(name))
}

func CheckLimitAndNotify(bot *tgbotapi.BotAPI, chatID int64, categoryID int, categoryName string, lang string) {
	month := time.Now().Format("2006-01")

	var spent float64
	err := db.DB.QueryRow(`
        SELECT COALESCE(SUM(amount), 0)
        FROM expenses
        WHERE user_id = ? AND category_id = ? AND date LIKE ?`,
		chatID, categoryID, month+"%").Scan(&spent)
	if err != nil {
		return
	}

	var limitSum float64
	err = db.DB.QueryRow("SELECT limit_sum FROM categories WHERE id = ? AND user_id = ?",
		categoryID, chatID).Scan(&limitSum)
	if err != nil || limitSum <= 0 {
		return
	}

	percent := (spent / limitSum) * 100

	var msgText string
	sent := false

	if percent >= 100 {
		msgText = utils.FormatAmount(i18n.Tf("limit_is_overloaded", lang, utils.Title.String(categoryName), spent, limitSum), lang)
		sent = true
	} else if percent >= 80 {
		msgText = utils.FormatAmount(i18n.Tf("limit_warning", lang, math.Round(percent), utils.Title.String(categoryName), spent, limitSum), lang)
		sent = true
	}

	if sent {
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

func HandleDeleteCategory(chatID int64, categoryName string, lang string) string {
	name := strings.ToLower(strings.TrimSpace(categoryName))

	var categoryID int
	var limit float64
	err := db.DB.QueryRow("SELECT id, limit_sum FROM categories WHERE user_id = ? AND LOWER(name) = ?",
		chatID, name).Scan(&categoryID, &limit)

	if err == sql.ErrNoRows {
		return i18n.T("category_not_found2", lang)
	} else if err != nil {
		return i18n.T("error_found_category", lang)
	}

	state.SetTempData(chatID, fmt.Sprintf("%d", categoryID))
	state.SetState(chatID, state.ConfirmDeleteCategory)

	displayName := utils.Title.String(name)
	return utils.FormatAmount(i18n.Tf("confirm_delete", lang, displayName, limit), lang)
}

func ConfirmDelete(chatID int64, answer string, lang string) string {
	categoryIDStr, err := state.GetTempData(chatID)
	if err != nil {
		state.Clear(chatID)
		return i18n.T("error_old_data", lang)
	}

	if strings.ToLower(answer) != "да" && strings.ToLower(answer) != "yes" {
		state.Clear(chatID)
		return i18n.T("delete_cancelled", lang)
	}

	var categoryID int
	fmt.Sscanf(categoryIDStr, "%d", &categoryID)

	tx, err := db.DB.Begin()
	if err != nil {
		return i18n.T("error_delete", lang)
	}

	_, err = tx.Exec("DELETE FROM expenses WHERE category_id = ?", categoryID)
	if err != nil {
		tx.Rollback()
		return i18n.T("error_delete_expense", lang)
	}

	_, err = tx.Exec("DELETE FROM categories WHERE id = ?", categoryID)
	if err != nil {
		tx.Rollback()
		return i18n.T("error_delete_category", lang)
	}

	err = tx.Commit()
	if err != nil {
		return i18n.T("error_confirm_changes", lang)
	}

	state.Clear(chatID)
	return i18n.T("category_deleted", lang)
}

func AddIncome(chatID int64, input string, lang string) string {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return i18n.T("invalid_format2", lang)
	}

	source := strings.ToLower(strings.TrimSpace(parts[0]))
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		return i18n.T("error_summ", lang)
	}

	_, err = db.DB.Exec("INSERT INTO incomes (user_id, source, amount, date) VALUES (?, ?, ?, ?)",
		chatID, source, amount, time.Now().Format("2006-01-02"))
	if err != nil {
		return i18n.T("error_save_income", lang)
	}

	return utils.FormatAmount(i18n.Tf("add_income", lang, utils.Title.String(source), amount), lang)
}

func IsPotentialIncome(text string) bool {
	parts := strings.Fields(text)
	if len(parts) != 2 {
		return false
	}
	_, err := strconv.ParseFloat(parts[1], 64)
	return err == nil
}

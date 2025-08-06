package commands

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"personal-finance/db"
	"personal-finance/state"
	"strconv"
	"strings"
	"time"
)

func GetMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Трата"),
			tgbotapi.NewKeyboardButton("🆕 Добавить категорию"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚙️ Категории"),
			tgbotapi.NewKeyboardButton("🎯 Лимиты"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Аналитика"),
			tgbotapi.NewKeyboardButton("💸 Изменить лимит"),
		),
	)
}

func AddCategory(chatID int64, input string) string {
	parts := strings.Fields(input)
	if len(parts) != 3 {
		return "Используйте: /addcategory имя_категории лимит (например: /addcategory еда 3000)"
	}

	name := strings.ToLower(parts[1])
	limit, err := strconv.ParseFloat(parts[2], 64)
	if err != nil || limit <= 0 {
		return "Лимит должен быть положительным числом."
	}

	_, err = db.DB.Exec("INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)
	if err != nil {
		return "Ошибка при добавлении категории."
	}

	return fmt.Sprintf("✅ Категория '%s' добавлена с лимитом %.2f ₽", name, limit)
}

func AddExpense(chatID int64, input string) string {
	parts := strings.Fields(input)
	categoryInput := strings.ToLower(strings.TrimSpace(parts[0]))
	var amount float64
	fmt.Sscanf(parts[1], "%f", &amount)

	var categoryID int
	err := db.DB.QueryRow("SELECT id FROM categories WHERE user_id = ? AND LOWER(name) = ?",
		chatID, categoryInput).Scan(&categoryID)

	if err == sql.ErrNoRows {
		return "❌ Категория не найдена. Проверьте название или добавьте новую."
	} else if err != nil {
		return "❌ Ошибка при поиске категории."
	}

	_, err = db.DB.Exec("INSERT INTO expenses (user_id, category_id, amount, date) VALUES (?, ?, ?, ?)",
		chatID, categoryID, amount, time.Now().Format("2006-01-02"))
	if err != nil {
		return "❌ Ошибка при сохранении траты."
	}

	return fmt.Sprintf("✅ Трата добавлена: %s — %.2f ₽", strings.Title(categoryInput), amount)
}

func ListCategories(chatID int64) string {
	rows, err := db.DB.Query("SELECT name FROM categories WHERE user_id = ?", chatID)
	if err != nil {
		return "Ошибка загрузки категорий."
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}

	if len(categories) == 0 {
		return "Категории отсутствуют. Добавьте командой /addcategory имя_ категории лимит"
	}
	return "Категории:\n• " + strings.Join(categories, "\n• ")
}

func ListLimits(chatID int64) string {
	rows, err := db.DB.Query("SELECT name, limit_sum FROM categories WHERE user_id = ?", chatID)
	if err != nil {
		return "Ошибка загрузки лимитов."
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		var limit float64
		rows.Scan(&name, &limit)
		result = append(result, fmt.Sprintf("%s: %.2f ₽", name, limit))
	}

	return "Лимиты:\n" + strings.Join(result, "\n")
}

func GetAnalytics(chatID int64) string {
	month := time.Now().Format("2006-01")

	rows, err := db.DB.Query(`
        SELECT c.name, SUM(e.amount), c.limit_sum 
        FROM expenses e
        JOIN categories c ON e.category_id = c.id
        WHERE e.user_id = ? AND e.date LIKE ?
        GROUP BY c.name, c.limit_sum`, chatID, month+"%")
	if err != nil {
		return "Ошибка аналитики."
	}
	defer rows.Close()

	var report []string
	total := 0.0

	for rows.Next() {
		var name string
		var spent, limit float64
		rows.Scan(&name, &spent, &limit)
		total += spent
		status := "✅"
		if spent > limit {
			status = "❌ (лимит превышен)"
		}
		report = append(report, fmt.Sprintf("%s: %.2f ₽ / %.2f ₽ %s", name, spent, limit, status))
	}

	if len(report) == 0 {
		return "Трат за месяц нет."
	}

	return fmt.Sprintf("📊 Траты за %s:\n\n%s\n\n💸 Всего: %.2f ₽", month, strings.Join(report, "\n"), total)
}

// IsPotentialExpense проверяет, является ли текст попыткой ввода траты: "категория сумма"
func IsPotentialExpense(chatID int64, text string) bool {
	parts := strings.Fields(text)
	if len(parts) != 2 {
		return false
	}

	categoryName := strings.ToLower(strings.TrimSpace(parts[0]))
	amountStr := parts[1]

	// Проверяем, что вторая часть — число
	_, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return false
	}

	// Проверяем, существует ли такая категория у пользователя
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

func CreateCategory(chatID int64, name string, limit float64) string {
	_, err := db.DB.Exec("INSERT INTO categories (name, user_id, limit_sum) VALUES (?, ?, ?)",
		name, chatID, limit)
	if err != nil {
		return "❌ Ошибка при создании категории."
	}
	return fmt.Sprintf("✅ Категория '%s' добавлена с лимитом %.2f ₽", strings.Title(name), limit)
}

func UpdateLimit(chatID int64, categoryName string, newLimit float64) string {
	result, err := db.DB.Exec("UPDATE categories SET limit_sum = ? WHERE user_id = ? AND LOWER(name) = ?",
		newLimit, chatID, strings.ToLower(categoryName))
	if err != nil {
		return "❌ Ошибка при обновлении лимита."
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return "❌ Категория не найдена."
	}

	return fmt.Sprintf("✅ Лимит для '%s' обновлён: %.2f ₽", strings.Title(categoryName), newLimit)
}

func HandleNewCategoryName(chatID int64, text string) string {
	name := strings.ToLower(strings.TrimSpace(text))
	if name == "" {
		return "Имя не может быть пустым. Попробуйте снова:"
	}

	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE user_id = ? AND LOWER(name) = ?", chatID, name).Scan(&count)
	if err != nil {
		return "Ошибка проверки."
	}
	if count > 0 {
		state.Clear(chatID)
		return "Категория уже существует."
	}

	// Сохраняем во временное хранилище Redis
	state.SetTempData(chatID, name)
	state.SetState(chatID, state.AwaitingCategoryLimit)

	return fmt.Sprintf("Категория: %s. Установите лимит:", strings.Title(name))
}

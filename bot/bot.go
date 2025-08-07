package bot

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"personal-finance/commands"
	"personal-finance/state"
	"strconv"
	"strings"
)

func StartBot(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go handleUpdate(bot, update)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	text := update.Message.Text
	chatID := update.Message.Chat.ID

	// Получаем текущее состояние
	userState, _ := state.GetState(chatID)

	// === 1. Сначала проверяем состояния ввода (FSM) ===
	switch userState {
	case state.AwaitingCategoryName:
		result := commands.HandleNewCategoryName(chatID, text)
		msg.Text = result
		bot.Send(msg)
		return

	case state.AwaitingCategoryLimit:
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = "Введите корректное положительное число."
			bot.Send(msg)
			return
		}
		categoryName, _ := state.GetTempData(chatID)
		msg.Text = commands.CreateCategory(chatID, categoryName, amount)
		state.Clear(chatID)
		bot.Send(msg)
		return

	case state.AwaitingLimitUpdate:
		// Сохраняем имя категории
		categoryName := strings.TrimSpace(text)
		if categoryName == "" {
			msg.Text = "Имя не может быть пустым. Попробуйте снова:"
			bot.Send(msg)
			return
		}
		state.SetTempData(chatID, categoryName)
		state.SetState(chatID, state.AwaitingNewLimitValue)
		msg.Text = fmt.Sprintf("Введите новый лимит для *%s*:", strings.Title(categoryName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return

	case state.AwaitingNewLimitValue:
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = "Введите корректное положительное число."
			bot.Send(msg)
			return
		}
		categoryName, err := state.GetTempData(chatID)
		if err != nil {
			msg.Text = "❌ Сессия устарела. Начните заново."
			state.Clear(chatID)
			bot.Send(msg)
			return
		}
		msg.Text = commands.UpdateLimit(chatID, categoryName, amount)
		state.Clear(chatID)
		bot.Send(msg)
		return

	case state.AwaitingCategoryToDelete:
		result := commands.HandleDeleteCategory(chatID, text)
		msg.Text = result
		bot.Send(msg)
		return

	case state.ConfirmDeleteCategory:
		result := commands.ConfirmDelete(chatID, text)
		msg.Text = result
		bot.Send(msg)
		return
	}

	// === 2. Обработка навигации (подменю) ===
	switch text {
	case "⬅️ Назад":
		state.SetState(chatID, state.MainMenu)
		msg.Text = "Главное меню:"
		msg.ReplyMarkup = commands.GetMainMenu()
		bot.Send(msg)
		return
	}

	// === 3. Обработка основных команд ===
	switch userState {
	case state.CategoriesMenu:
		handleCategoriesMenu(chatID, text, &msg)
		bot.Send(msg)
		return

	case state.LimitsMenu:
		handleLimitsMenu(chatID, text, &msg)
		bot.Send(msg)
		return
	}

	// === 4. Обычные команды ===
	switch text {
	case "/start":
		state.SetState(chatID, state.MainMenu)
		msg.Text = "Привет! Я финансовый помощник."
		msg.ReplyMarkup = commands.GetMainMenu()

	case "📊 Аналитика":
		msg.Text = commands.GetAnalytics(chatID)
		msg.ReplyMarkup = commands.GetMainMenu()

	case "💵 Доход":
		msg.Text = "Введите: источник сумма (например: зарплата 100000)"
		msg.ReplyMarkup = commands.GetMainMenu()

	case "➕ Трата":
		msg.Text = "Введите: категория сумма (например: еда 500)"
		msg.ReplyMarkup = commands.GetMainMenu()

	case "⚙️ Категории":
		state.SetState(chatID, state.CategoriesMenu)
		msg.Text = "🔧 Управление категориями:"
		msg.ReplyMarkup = commands.GetCategoriesMenu()

	case "🎯 Лимиты":
		state.SetState(chatID, state.LimitsMenu)
		msg.Text = "🎯 Управление лимитами:"
		msg.ReplyMarkup = commands.GetLimitsMenu()

	default:
		if commands.IsPotentialExpense(chatID, text) {
			msg.Text = commands.AddExpense(bot, chatID, text)
		} else if isPotentialIncome(text) {
			msg.Text = commands.AddIncome(chatID, text)
		} else {
			msg.Text = "Неизвестная команда."
			msg.ReplyMarkup = commands.GetMainMenu()
		}
	}

	bot.Send(msg)
}

func isPotentialIncome(text string) bool {
	parts := strings.Fields(text)
	if len(parts) != 2 {
		return false
	}
	_, err := strconv.ParseFloat(parts[1], 64)
	return err == nil
}

func handleCategoriesMenu(chatID int64, text string, msg *tgbotapi.MessageConfig) {
	switch text {
	case "📋 Список категорий":
		msg.Text = commands.ListCategories(chatID)
		msg.ReplyMarkup = commands.GetCategoriesMenu()

	case "➕ Добавить категорию":
		msg.Text = "Введите название новой категории:"
		state.SetState(chatID, state.AwaitingCategoryName)
		msg.ReplyMarkup = commands.GetCategoriesMenu()

	case "🗑 Удалить категорию":
		msg.Text = "Введите название категории для удаления:"
		state.SetState(chatID, state.AwaitingCategoryToDelete)
		msg.ReplyMarkup = commands.GetCategoriesMenu()

	case "⬅️ Назад":
		state.SetState(chatID, state.MainMenu)
		msg.Text = "Главное меню:"
		msg.ReplyMarkup = commands.GetMainMenu()

	default:
		msg.Text = "Неизвестная команда. Используйте кнопки."
		msg.ReplyMarkup = commands.GetCategoriesMenu()
	}
}

func handleLimitsMenu(chatID int64, text string, msg *tgbotapi.MessageConfig) {
	switch text {
	case "🎯 Лимиты по категориям":
		msg.Text = commands.ListLimits(chatID)
		msg.ReplyMarkup = commands.GetLimitsMenu()

	case "💸 Изменить лимит":
		msg.Text = "Введите название категории:"
		state.SetState(chatID, state.AwaitingLimitUpdate)
		msg.ReplyMarkup = commands.GetLimitsMenu()

	case "⬅️ Назад":
		state.SetState(chatID, state.MainMenu)
		msg.Text = "Главное меню:"
		msg.ReplyMarkup = commands.GetMainMenu()

	default:
		msg.Text = "Неизвестная команда. Используйте кнопки."
		msg.ReplyMarkup = commands.GetLimitsMenu()
	}
}

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

	// Получаем состояние из Redis
	userState, _ := state.GetState(chatID)

	switch userState {
	case state.AwaitingCategoryName:
		msg.Text = commands.HandleNewCategoryName(chatID, text)
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
		state.Clear(chatID) // Очистка Redis
		bot.Send(msg)
		return

	case state.AwaitingLimitUpdate:
		// Пользователь вводит название категории
		categoryName := strings.TrimSpace(text)
		if categoryName == "" {
			msg.Text = "Имя не может быть пустым. Попробуйте снова:"
			bot.Send(msg)
			return
		}

		// Сохраняем название категории
		state.SetTempData(chatID, categoryName)
		state.SetState(chatID, state.AwaitingNewLimitValue)

		msg.Text = fmt.Sprintf("Введите новый лимит для категории *%s*:", strings.Title(categoryName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return

	case state.AwaitingNewLimitValue:
		// Пользователь вводит лимит
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = "Введите корректное положительное число."
			bot.Send(msg)
			return
		}

		// Получаем сохранённое имя категории
		categoryName, err := state.GetTempData(chatID)
		if err != nil {
			msg.Text = "❌ Сессия устарела. Начните заново."
			state.Clear(chatID)
			bot.Send(msg)
			return
		}

		// Выполняем обновление
		msg.Text = commands.UpdateLimit(chatID, categoryName, amount)
		state.Clear(chatID) // сброс состояния
		bot.Send(msg)
		return

	case state.AwaitingCategoryToDelete:
		msg.Text = commands.HandleDeleteCategory(chatID, text)
		bot.Send(msg)
		return

	case state.ConfirmDeleteCategory:
		msg.Text = commands.ConfirmDelete(chatID, text)
		bot.Send(msg)
		return
	}

	// Обычные команды
	switch {
	case text == "/start":
		state.Clear(chatID) // сброс
		msg.Text = "Привет! Я финансовый помощник."
		msg.ReplyMarkup = commands.GetMainKeyboard()

	case text == "📊 Аналитика":
		msg.Text = commands.GetAnalytics(chatID)

	case text == "➕ Трата":
		msg.Text = "Введите: категория сумма (например: еда 500)"

	case text == "💵 Доход":
		msg.Text = "Введите: источник сумма (например: зарплата 100000)"

	case text == "⚙️ Категории":
		msg.Text = commands.ListCategories(chatID)

	case text == "🎯 Лимиты":
		msg.Text = commands.ListLimits(chatID)

	case text == "🆕 Добавить категорию":
		msg.Text = "Введите название новой категории:"
		state.SetState(chatID, state.AwaitingCategoryName)

	case text == "💸 Изменить лимит":
		msg.Text = "Введите название категории:"
		state.SetState(chatID, state.AwaitingLimitUpdate)

	case text == "🗑 Удалить категорию":
		msg.Text = "Введите название категории, которую хотите удалить:"
		state.SetState(chatID, state.AwaitingCategoryToDelete)

	default:
		if commands.IsPotentialExpense(chatID, text) {
			msg.Text = commands.AddExpense(bot, chatID, text)
		} else if isPotentialIncome(text) {
			msg.Text = commands.AddIncome(chatID, text)
		} else {
			msg.Text = "Неизвестная команда. Используйте меню."
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
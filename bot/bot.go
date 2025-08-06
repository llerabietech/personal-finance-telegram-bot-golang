package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"personal-finance/commands"
	"personal-finance/state"
	"strconv"
)

func StartBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

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
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = "Введите корректное положительное число."
			bot.Send(msg)
			return
		}
		categoryName, _ := state.GetTempData(chatID)
		msg.Text = commands.UpdateLimit(chatID, categoryName, amount)
		state.Clear(chatID)
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

	default:
		if commands.IsPotentialExpense(chatID, text) {
			msg.Text = commands.AddExpense(bot, chatID, text)
		} else {
			msg.Text = "Неизвестная команда. Используйте меню."
		}
	}

	bot.Send(msg)
}

package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"personal-finance/commands"
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

	switch {
	case text == "/start":
		msg.Text = "Привет! Я финансовый помощник. Используй команды:"
		msg.ReplyMarkup = commands.GetMainKeyboard()
	case text == "📊 Аналитика":
		msg.Text = commands.GetAnalytics(chatID)
	case text == "➕ Трата":
		msg.Text = "Введите: категория сумма (например: еда 500)"
	case text == "⚙️ Категории":
		msg.Text = commands.ListCategories(chatID)
	case text == "🎯 Лимиты":
		msg.Text = commands.ListLimits(chatID)
	case strings.HasPrefix(text, "/addcategory"):
    	msg.Text = commands.AddCategory(chatID, text)
	default:
		// Обработка ввода траты
		if commands.IsPotentialExpense(chatID, text) {
			response := commands.AddExpense(chatID, text)
			msg.Text = response
		} else {
			msg.Text = "Неизвестная команда. Используйте меню."
		}
	}

	bot.Send(msg)
}

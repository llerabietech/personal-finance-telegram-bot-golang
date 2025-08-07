package bot

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"personal-finance/commands"
	"personal-finance/i18n"
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

	// Получаем язык
	lang, err := state.GetUserLanguage(chatID)
	if err != nil {
		lang = "en"
	}

	// Получаем текущее состояние
	userState, _ := state.GetState(chatID)

	// === 1. Сначала проверяем состояния ввода (FSM) ===
	switch userState {
	case state.AwaitingCategoryName:
		result := commands.HandleNewCategoryName(chatID, text, lang)
		msg.Text = result
		bot.Send(msg)
		return

	case state.AwaitingCategoryLimit:
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = i18n.T("invalid_amount", lang)
			bot.Send(msg)
			return
		}
		categoryName, _ := state.GetTempData(chatID)
		msg.Text = commands.CreateCategory(chatID, categoryName, amount, lang)
		state.Clear(chatID)
		msg.ReplyMarkup = commands.GetMainMenu(lang)
		bot.Send(msg)
		return

	case state.AwaitingLimitUpdate:
		// Сохраняем имя категории
		categoryName := strings.TrimSpace(text)
		if categoryName == "" {
			msg.Text = i18n.T("empty_name", lang)
			bot.Send(msg)
			return
		}
		state.SetTempData(chatID, categoryName)
		state.SetState(chatID, state.AwaitingNewLimitValue)
		msg.Text = fmt.Sprintf(i18n.T("enter_new_limit", lang), strings.Title(categoryName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return

	case state.AwaitingNewLimitValue:
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = i18n.T("correct_digit", lang)
			bot.Send(msg)
			return
		}
		categoryName, err := state.GetTempData(chatID)
		if err != nil {
			msg.Text = i18n.T("outdated_session", lang)
			state.Clear(chatID)
			bot.Send(msg)
			return
		}
		msg.Text = commands.UpdateLimit(chatID, categoryName, amount, lang)
		state.Clear(chatID)
		bot.Send(msg)
		return

	case state.AwaitingCategoryToDelete:
		result := commands.HandleDeleteCategory(chatID, text, lang)
		msg.Text = result
		bot.Send(msg)
		return

	case state.ConfirmDeleteCategory:
		result := commands.ConfirmDelete(chatID, text, lang)
		msg.Text = result
		bot.Send(msg)
		return
	}

	// === 2. Обработка команд ===
	switch text {
	case "/start":
		state.SetState(chatID, state.StateChoosingLanguage)
		msg.Text = i18n.T("welcome", lang)
		msg.ReplyMarkup = commands.GetLanguageKeyboard()
		bot.Send(msg)
		return

	case "🇷🇺 Русский", "🇬🇧 English":
		if userState == state.StateChoosingLanguage {
			selectedLang := "ru"
			if text == "🇬🇧 English" {
				selectedLang = "en"
			}
			state.SetUserLanguage(chatID, selectedLang)
			state.SetState(chatID, state.StateMainMenu)

			msg.Text = i18n.T("language_selected", selectedLang)
			msg.ReplyMarkup = commands.GetMainMenu(selectedLang)
			bot.Send(msg)
		} else {
			msg.Text = i18n.T("menu_main", lang)
			msg.ReplyMarkup = commands.GetMainMenu(lang)
			bot.Send(msg)
		}
		return

	case i18n.T("back", lang):
		state.SetState(chatID, state.StateMainMenu)
		msg.Text = i18n.T("menu_main", lang)
		msg.ReplyMarkup = commands.GetMainMenu(lang)
		bot.Send(msg)
		return
	}

	// === 3. Подменю и команды ===
	switch userState {
	case state.CategoriesMenu:
		handleCategoriesMenu(chatID, text, &msg, lang)
		bot.Send(msg)
		return
	case state.LimitsMenu:
		handleLimitsMenu(chatID, text, &msg, lang)
		bot.Send(msg)
		return
	}

	// === 4. Основные команды ===
	switch text {
	case i18n.T("analytics", lang):
		msg.Text = commands.GetAnalytics(chatID, lang)
		msg.ReplyMarkup = commands.GetMainMenu(lang)

	case i18n.T("income", lang):
		msg.Text = i18n.T("enter_income", lang)
		msg.ReplyMarkup = commands.GetMainMenu(lang)

	case i18n.T("expenses", lang):
		msg.Text = i18n.T("enter_expense", lang)
		msg.ReplyMarkup = commands.GetMainMenu(lang)

	case i18n.T("categories", lang):
		state.SetState(chatID, state.CategoriesMenu)
		msg.Text = "🔧 " + i18n.T("categories", lang)
		msg.ReplyMarkup = commands.GetCategoriesMenu(lang)

	case i18n.T("limits", lang):
		state.SetState(chatID, state.LimitsMenu)
		msg.Text = i18n.T("limits", lang)
		msg.ReplyMarkup = commands.GetLimitsMenu(lang)

	default:
		if commands.IsPotentialExpense(chatID, text) {
			msg.Text = commands.AddExpense(bot, chatID, text, lang)
			msg.ReplyMarkup = commands.GetMainMenu(lang)
		} else if isPotentialIncome(text) {
			msg.Text = commands.AddIncome(chatID, text, lang)
			msg.ReplyMarkup = commands.GetMainMenu(lang)
		} else {
			msg.Text = "❌ " + i18n.T("invalid_format", lang)
			msg.ReplyMarkup = commands.GetMainMenu(lang)
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

func handleCategoriesMenu(chatID int64, text string, msg *tgbotapi.MessageConfig, lang string) {
	switch text {
	case i18n.T("list_categories", lang):
		msg.Text = commands.ListCategories(chatID, lang)
		msg.ReplyMarkup = commands.GetCategoriesMenu(lang)

	case i18n.T("add_category", lang):
		msg.Text = i18n.T("enter_category_name", lang)
		state.SetState(chatID, state.AwaitingCategoryName)
		msg.ReplyMarkup = commands.GetCategoriesMenu(lang)

	case i18n.T("delete_category", lang):
		msg.Text = i18n.T("enter_category_name_delete", lang)
		state.SetState(chatID, state.AwaitingCategoryToDelete)
		msg.ReplyMarkup = commands.GetCategoriesMenu(lang)

	case i18n.T("back", lang):
		state.SetState(chatID, state.MainMenu)
		msg.Text = i18n.T("menu_main", lang)
		msg.ReplyMarkup = commands.GetMainMenu(lang)

	default:
		msg.Text = i18n.T("unknown_cmd", lang)
		msg.ReplyMarkup = commands.GetCategoriesMenu(lang)
	}
}

func handleLimitsMenu(chatID int64, text string, msg *tgbotapi.MessageConfig, lang string) {
	switch text {
	case i18n.T("limits_list", lang):
		msg.Text = commands.ListLimits(chatID, lang)
		msg.ReplyMarkup = commands.GetLimitsMenu(lang)

	case i18n.T("change_limit", lang):
		msg.Text = i18n.T("enter_category_name_for_limit", lang)
		state.SetState(chatID, state.AwaitingLimitUpdate)
		msg.ReplyMarkup = commands.GetLimitsMenu(lang)

	case i18n.T("back", lang):
		state.SetState(chatID, state.MainMenu)
		msg.Text = i18n.T("menu_main", lang)
		msg.ReplyMarkup = commands.GetMainMenu(lang)

	default:
		msg.Text = i18n.T("unknown_cmd", lang)
		msg.ReplyMarkup = commands.GetLimitsMenu(lang)
	}
}

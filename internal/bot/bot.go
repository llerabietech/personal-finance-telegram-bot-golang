package bot

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/config"
	"personal-finance/internal/helper"
	"personal-finance/internal/i18n"
	"personal-finance/internal/service"
	"personal-finance/internal/ui"
	"personal-finance/state"
	"personal-finance/utils"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(bot *tgbotapi.BotAPI, db *sql.DB, redis *redis.Client, cfg *config.Config) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go handleUpdate(bot, update, db, redis, cfg)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, redis *redis.Client, cfg *config.Config) {
	ctx := context.Background()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	text := update.Message.Text
	chatID := update.Message.Chat.ID

	lang, err := state.GetUserLanguage(ctx, redis, chatID)
	if err != nil {
		lang = "en"
	}

	userState, _ := state.GetState(ctx, redis, chatID)

	// check the input states
	switch userState {
	case state.AwaitingCategoryName:
		result := service.NewCategoryName(ctx, db, redis, chatID, text, lang, cfg)
		msg.Text = result
		bot.Send(msg)
		return

	case state.AwaitingCategoryLimit:
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = i18n.T("invalid_amount", lang, cfg)
			bot.Send(msg)
			return
		}
		categoryName, _ := state.GetTempData(ctx, redis, chatID)
		msg.Text = service.CreateCategory(ctx, db, chatID, categoryName, amount, lang, cfg)
		state.Clear(ctx, redis, chatID)
		msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)
		bot.Send(msg)
		return

	case state.AwaitingLimitUpdate:
		categoryName := strings.TrimSpace(text)
		if categoryName == "" {
			msg.Text = i18n.T("empty_name", lang, cfg)
			bot.Send(msg)
			return
		}
		state.SetTempData(ctx, redis, chatID, categoryName)
		state.SetState(ctx, redis, chatID, state.AwaitingNewLimitValue)
		msg.Text = fmt.Sprintf(i18n.T("enter_new_limit", lang, cfg), utils.Title.String(categoryName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return

	case state.AwaitingNewLimitValue:
		amount, err := strconv.ParseFloat(text, 64)
		if err != nil || amount <= 0 {
			msg.Text = i18n.T("correct_digit", lang, cfg)
			bot.Send(msg)
			return
		}
		categoryName, err := state.GetTempData(ctx, redis, chatID)
		if err != nil {
			msg.Text = i18n.T("outdated_session", lang, cfg)
			state.Clear(ctx, redis, chatID)
			bot.Send(msg)
			return
		}
		msg.Text = service.UpdateLimit(ctx, db, chatID, categoryName, amount, lang, cfg)
		state.Clear(ctx, redis, chatID)
		bot.Send(msg)
		return

	case state.AwaitingCategoryToDelete:
		result := service.DeleteCategory(ctx, db, redis, chatID, text, lang, cfg)
		msg.Text = result
		bot.Send(msg)
		return

	case state.ConfirmDeleteCategory:
		result := helper.ConfirmDelete(ctx, db, redis, chatID, text, lang, cfg)
		msg.Text = result
		bot.Send(msg)
		return
	}

	if userState == state.StateChoosingLanguage {
		if code, ok := i18n.DetectLanguageFromButton(text, cfg); ok {
			state.SetUserLanguage(ctx, redis, chatID, code)
			state.SetState(ctx, redis, chatID, state.StateMainMenu)

			msg.Text = i18n.T("language_selected", code, cfg)
			msg.ReplyMarkup = ui.GetMainMenu(code, cfg)
			bot.Send(msg)
			return
		}
	}

	// Command processing
	switch text {
	case "/start":
		state.SetState(ctx, redis, chatID, state.StateChoosingLanguage)
		msg.Text = i18n.T("welcome", lang, cfg)
		msg.ReplyMarkup = ui.GetLanguageKeyboard(cfg)
		bot.Send(msg)
		return

	case "🇷🇺 Русский", "🇬🇧 English":
		if userState == state.StateChoosingLanguage {
			selectedLang := ""
			if code, ok := i18n.DetectLanguageFromButton(text, cfg); ok {
				selectedLang = code
			} else {
				selectedLang = cfg.App.DefaultLanguage
			}
			state.SetUserLanguage(ctx, redis, chatID, selectedLang)
			state.SetState(ctx, redis, chatID, state.StateMainMenu)

			msg.Text = i18n.T("language_selected", selectedLang, cfg)
			msg.ReplyMarkup = ui.GetMainMenu(selectedLang, cfg)
			bot.Send(msg)
		} else {
			msg.Text = i18n.T("menu_main", lang, cfg)
			msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)
			bot.Send(msg)
		}
		return

	case i18n.T("back", lang, cfg):
		state.SetState(ctx, redis, chatID, state.StateMainMenu)
		msg.Text = i18n.T("menu_main", lang, cfg)
		msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)
		bot.Send(msg)
		return
	}

	// Submenus and commands
	switch userState {
	case state.CategoriesMenu:
		handleCategoriesMenu(ctx, db, redis, chatID, text, &msg, lang, cfg)
		bot.Send(msg)
		return
	case state.LimitsMenu:
		handleLimitsMenu(ctx, db, redis, chatID, text, &msg, lang, cfg)
		bot.Send(msg)
		return
	}

	// Basic commands
	switch text {
	case i18n.T("analytics", lang, cfg):
		msg.Text = service.GetAnalytics(ctx, db, chatID, lang, cfg)
		msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)

	case i18n.T("income", lang, cfg):
		msg.Text = i18n.T("enter_income", lang, cfg)
		msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)

	case i18n.T("expenses", lang, cfg):
		msg.Text = i18n.T("enter_expense", lang, cfg)
		msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)

	case i18n.T("categories", lang, cfg):
		state.SetState(ctx, redis, chatID, state.CategoriesMenu)
		msg.Text = i18n.T("categories", lang, cfg)
		msg.ReplyMarkup = ui.GetCategoriesMenu(lang, cfg)

	case i18n.T("limits", lang, cfg):
		state.SetState(ctx, redis, chatID, state.LimitsMenu)
		msg.Text = i18n.T("limits", lang, cfg)
		msg.ReplyMarkup = ui.GetLimitsMenu(lang, cfg)

	default:
		if service.IsPotentialExpense(ctx, db, chatID, text) {
			msg.Text = service.AddExpense(bot, ctx, db, chatID, text, lang, cfg)
			msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)
		} else if service.IsPotentialIncome(ctx, db, chatID, text) {
			msg.Text = service.AddIncome(ctx, db, chatID, text, lang, cfg)
			msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)
		} else {
			msg.Text = i18n.T("invalid_format", lang, cfg)
			msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)
		}
	}

	bot.Send(msg)
}

func handleCategoriesMenu(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, text string, msg *tgbotapi.MessageConfig, lang string, cfg *config.Config) {
	switch text {
	case i18n.T("list_categories", lang, cfg):
		msg.Text = service.ListCategories(ctx, db, chatID, lang, cfg)
		msg.ReplyMarkup = ui.GetCategoriesMenu(lang, cfg)

	case i18n.T("add_category", lang, cfg):
		msg.Text = i18n.T("enter_category_name", lang, cfg)
		state.SetState(ctx, redis, chatID, state.AwaitingCategoryName)
		msg.ReplyMarkup = ui.GetCategoriesMenu(lang, cfg)

	case i18n.T("delete_category", lang, cfg):
		msg.Text = i18n.T("enter_category_name_delete", lang, cfg)
		state.SetState(ctx, redis, chatID, state.AwaitingCategoryToDelete)
		msg.ReplyMarkup = ui.GetCategoriesMenu(lang, cfg)

	case i18n.T("back", lang, cfg):
		state.SetState(ctx, redis, chatID, state.MainMenu)
		msg.Text = i18n.T("menu_main", lang, cfg)
		msg.ReplyMarkup = ui.GetMainMenu(lang, cfg)

	default:
		msg.Text = i18n.T("unknown_cmd", lang, cfg)
		msg.ReplyMarkup = ui.GetCategoriesMenu(lang, cfg)
	}
}

func handleLimitsMenu(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, text string, msg *tgbotapi.MessageConfig, lang string, cfg *config.Config) {
	switch text {
	case i18n.T("limits_list", lang, cfg):
		msg.Text = service.ListLimits(ctx, db, chatID, lang, cfg)
		msg.ReplyMarkup = ui.GetLimitsMenu(lang, cfg)

	case i18n.T("change_limit", lang, cfg):
		msg.Text = i18n.T("enter_category_name_for_limit", lang, cfg)
		state.SetState(ctx, redis, chatID, state.AwaitingLimitUpdate)
		msg.ReplyMarkup = ui.GetLimitsMenu(lang, cfg)

	case i18n.T("back", lang, cfg):
		state.SetState(ctx, redis, chatID, state.MainMenu)
		msg.Text = i18n.T("menu_main", lang, cfg)
		msg.ReplyMarkup = ui.GetLimitsMenu(lang, cfg)

	default:
		msg.Text = i18n.T("unknown_cmd", lang, cfg)
		msg.ReplyMarkup = ui.GetLimitsMenu(lang, cfg)
	}
}

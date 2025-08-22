package ui

import (
	"personal-finance/internal/config"
	"personal-finance/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetLanguageKeyboard(cfg *config.Config) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	currentRow := []tgbotapi.KeyboardButton{}
	for _, code := range cfg.App.Languages {
		label := i18n.LanguageButton(code)
		currentRow = append(currentRow, tgbotapi.NewKeyboardButton(label))
		if len(currentRow) == 2 {
			rows = append(rows, tgbotapi.NewKeyboardButtonRow(currentRow...))
			currentRow = []tgbotapi.KeyboardButton{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(currentRow...))
	}
	return tgbotapi.NewReplyKeyboard(rows...)
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

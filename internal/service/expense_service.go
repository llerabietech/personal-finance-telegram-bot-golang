package service

import (
	"context"
	"database/sql"
	"math"
	"personal-finance/internal/config"
	"personal-finance/internal/i18n"
	"personal-finance/internal/repository"
	"personal-finance/utils"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

	categoryID, categoryName, err := repository.GetCategory(ctx, db, chatID, categoryInput)

	if err == sql.ErrNoRows {
		return i18n.Tf("category_not_found", lang, cfg, categoryInput)
	} else if err != nil {
		return i18n.T("error_found_category", lang, cfg)
	}

	err = repository.AddExpense(ctx, db, chatID, categoryID, amount, cfg)
	if err != nil {
		return i18n.T("error_save_expense", lang, cfg)
	}

	go CheckLimitAndNotify(bot, ctx, db, chatID, categoryID, categoryName, lang, cfg)

	return utils.FormatAmount(i18n.Tf("add_expense", lang, cfg, utils.Title.String(categoryName), amount), lang, cfg)
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

	return repository.IsPotentialExpense(ctx, db, chatID, categoryName)
}

func CheckLimitAndNotify(bot *tgbotapi.BotAPI, ctx context.Context, db *sql.DB, chatID int64, categoryID int, categoryName string, lang string, cfg *config.Config) {
	month := time.Now().Format(cfg.App.MonthFormat)

	spent, err := repository.GetSpent(ctx, db, chatID, categoryID, month)
	if err != nil {
		return
	}

	limitSum, err := repository.GetLimitSum(ctx, db, chatID, categoryID)

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
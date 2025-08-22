package commands

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/i18n"
	"personal-finance/internal/config"
	"personal-finance/state"
	"personal-finance/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)



func ConfirmDelete(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, answer string, lang string, cfg *config.Config) string {
	categoryIDStr, err := state.GetTempData(ctx, redis, chatID)
	if err != nil {
		state.Clear(ctx, redis, chatID)
		return i18n.T("error_old_data", lang, cfg)
	}

	// Проверяем, является ли ответ одним из подтверждающих слов
	isConfirmed := false
	for _, word := range cfg.App.ConfirmationWords {
		if strings.EqualFold(answer, word){
			isConfirmed = true
			break
		}
	}
	if !isConfirmed {
		state.Clear(ctx, redis, chatID)
		return i18n.T("delete_cancelled", lang, cfg)
	}

	var categoryID int
	fmt.Sscanf(categoryIDStr, "%d", &categoryID)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return i18n.T("error_delete", lang, cfg)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM expenses WHERE category_id = ?", categoryID)
	if err != nil {
		tx.Rollback()
		return i18n.T("error_delete_expense", lang, cfg)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", categoryID)
	if err != nil {
		tx.Rollback()
		return i18n.T("error_delete_category", lang, cfg)
	}

	err = tx.Commit()
	if err != nil {
		return i18n.T("error_confirm_changes", lang, cfg)
	}

	state.Clear(ctx, redis, chatID)
	return i18n.T("category_deleted", lang, cfg)
}

func AddIncome(ctx context.Context, db *sql.DB, chatID int64, input string, lang string, cfg *config.Config) string {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return i18n.T("invalid_format2", lang, cfg)
	}

	source := strings.ToLower(strings.TrimSpace(parts[0]))
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		return i18n.T("error_summ", lang, cfg)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO incomes (user_id, source, amount, date) VALUES (?, ?, ?, ?)",
		chatID, source, amount, time.Now().Format(cfg.App.DateFormat))
	if err != nil {
		return i18n.T("error_save_income", lang, cfg)
	}

	return utils.FormatAmount(i18n.Tf("add_income", lang, cfg, utils.Title.String(source), amount), lang, cfg)
}

func IsPotentialIncome(ctx context.Context, db *sql.DB, chatID int64, text string) bool {
	parts := strings.Fields(text)
	if len(parts) != 2 {
		return false
	}
	_, err := strconv.ParseFloat(parts[1], 64)
	return err == nil
}

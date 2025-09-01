package helper

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/config"
	"personal-finance/internal/i18n"
	"personal-finance/internal/repository"
	"personal-finance/state"
	"strings"

	"github.com/go-redis/redis/v8"
)

func ConfirmDelete(ctx context.Context, db *sql.DB, redis *redis.Client, chatID int64, answer string, lang string, cfg *config.Config) string {
	categoryIDStr, err := state.GetTempData(ctx, redis, chatID)
	if err != nil {
		state.Clear(ctx, redis, chatID)
		return i18n.T("error_old_data", lang, cfg)
	}

	// check confirmation words
	isConfirmed := false
	for _, word := range cfg.App.ConfirmationWords {
		if strings.EqualFold(answer, word) {
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

	if err := repository.DeleteCategoryCascadeByID(ctx, db, categoryID); err != nil {
		switch err {
		case repository.ErrBeginTx:
			return i18n.T("error_delete", lang, cfg)
		case repository.ErrDeleteExpenses:
			return i18n.T("error_delete_expense", lang, cfg)
		case repository.ErrDeleteCategory:
			return i18n.T("error_delete_category", lang, cfg)
		case repository.ErrCommitTx:
			return i18n.T("error_confirm_changes", lang, cfg)
		default:
			return i18n.T("error_delete", lang, cfg)
		}
	}

	state.Clear(ctx, redis, chatID)
	return i18n.T("category_deleted", lang, cfg)
}

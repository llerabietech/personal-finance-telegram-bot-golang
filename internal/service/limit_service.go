package service

import (
	"context"
	"database/sql"
	"fmt"
	"personal-finance/internal/config"
	"personal-finance/internal/i18n"
	"personal-finance/internal/repository"
	"personal-finance/utils"
	"strings"
)

func ListLimits(ctx context.Context, db *sql.DB, chatID int64, lang string, cfg *config.Config) string {
	rows, err := repository.GetLimits(ctx, db, chatID)
	if err != nil {
		return i18n.T("error_load_limits", lang, cfg)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		var limit float64
		rows.Scan(&name, &limit)
		result = append(result, fmt.Sprintf("%s: %.2f %s", name, limit, cfg.App.CurrencySymbol))
	}

	return i18n.T("limits2", lang, cfg) + strings.Join(result, "\n")
}

func UpdateLimit(ctx context.Context, db *sql.DB, chatID int64, categoryName string, newLimit float64, lang string, cfg *config.Config) string {
	name := strings.ToLower(strings.TrimSpace(categoryName))
	if name == "" {
		return i18n.T("error_category_name_is_empty", lang, cfg)
	}

	categoryID, currentLimit, err := repository.GetCurrentLimit(ctx, db, chatID, name)

	if err == sql.ErrNoRows {
		return i18n.Tf("category_not_found", lang, cfg, utils.Title.String(name))
	} else if err != nil {
		return i18n.T("error_found_category", lang, cfg) + err.Error()
	}

	err = repository.UpdateLimit(ctx, db, newLimit, categoryID)
	if err != nil {
		return i18n.T("error_update_limit", lang, cfg)
	}

	return utils.FormatAmount(i18n.Tf("updated_limit", lang, cfg, utils.Title.String(name), currentLimit, newLimit), lang, cfg)
}

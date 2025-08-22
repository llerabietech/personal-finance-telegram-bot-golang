package commands

import (
	"context"
	"database/sql"
	"personal-finance/internal/i18n"
	"personal-finance/internal/config"
	"personal-finance/utils"
	"strconv"
	"strings"
	"time"
)





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
